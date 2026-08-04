package gitmirror

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/wavefnd/wave-platform/internal/sourceanalysis"
	"github.com/wavefnd/wave-platform/internal/storage"
)

const syncInterval = 15 * time.Minute

type Service struct {
	database     *storage.Database
	git          *gitRepository
	repositories map[string]Repository
	locks        map[string]*sync.Mutex
	stateMu      sync.RWMutex
	analyzer     sourceanalysis.Analyzer
}

func NewService(database *storage.Database, analyzers ...sourceanalysis.Analyzer) (*Service, error) {
	definitions := OfficialRepositories()
	service := &Service{database: database, git: newGitRepository(filepath.Join(database.Root, "git", "mirrors")), repositories: make(map[string]Repository, len(definitions)), locks: make(map[string]*sync.Mutex, len(definitions))}
	if len(analyzers) > 0 {
		service.analyzer = analyzers[0]
	}
	for _, definition := range definitions {
		stored, err := service.load(definition.ID)
		if err == nil {
			definition.LastFetchedAt = stored.LastFetchedAt
			definition.Status = stored.Status
			definition.HeadOID = stored.HeadOID
		} else if !errors.Is(err, storage.ErrNotFound) {
			return nil, fmt.Errorf("load Git repository %s: %w", definition.ID, err)
		}
		service.repositories[definition.ID] = definition
		service.locks[definition.ID] = &sync.Mutex{}
		if err := service.save(definition); err != nil {
			return nil, err
		}
	}
	return service, nil
}

func OfficialRepositories() []Repository {
	return []Repository{
		{ID: "wave", Owner: "wavefnd", Name: "Wave", Description: "Wave programming language compiler and standard library.", RemoteURL: "https://github.com/wavefnd/Wave.git", DefaultBranch: "master", Status: "pending"},
		{ID: "vex", Owner: "wavefnd", Name: "Vex", Description: "Package manager and build tool for Wave projects.", RemoteURL: "https://github.com/wavefnd/Vex.git", DefaultBranch: "master", Status: "pending"},
		{ID: "whale", Owner: "wavefnd", Name: "Whale", Description: "Assembler, object, linker, and IR toolchain for Wave.", RemoteURL: "https://github.com/wavefnd/Whale.git", DefaultBranch: "master", Status: "pending"},
	}
}

func (service *Service) Run(ctx context.Context) {
	service.SynchronizeAll(ctx)
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			service.SynchronizeAll(ctx)
		}
	}
}

func (service *Service) SynchronizeAll(ctx context.Context) {
	var wait sync.WaitGroup
	for id := range service.repositories {
		id := id
		wait.Add(1)
		go func() {
			defer wait.Done()
			if err := service.Synchronize(ctx, id); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("synchronize Git mirror %s: %v", id, err)
			}
		}()
	}
	wait.Wait()
}

func (service *Service) Synchronize(parent context.Context, id string) error {
	service.stateMu.RLock()
	definition, ok := service.repositories[id]
	service.stateMu.RUnlock()
	if !ok {
		return storage.ErrNotFound
	}
	lock := service.locks[id]
	lock.Lock()
	defer lock.Unlock()

	ctx, cancel := withTimeout(parent, 5*time.Minute)
	defer cancel()
	definition.Status = "syncing"
	service.stateMu.Lock()
	service.repositories[id] = definition
	service.stateMu.Unlock()
	_ = service.save(definition)
	if err := service.git.sync(ctx, definition); err != nil {
		definition.Status = "error"
		service.stateMu.Lock()
		service.repositories[id] = definition
		service.stateMu.Unlock()
		_ = service.save(definition)
		return err
	}
	branch, err := service.git.defaultBranch(ctx, id)
	if err == nil && branch != "" {
		definition.DefaultBranch = branch
	}
	head, err := service.git.resolve(ctx, id, "refs/heads/"+definition.DefaultBranch)
	if err != nil {
		definition.Status = "error"
		service.stateMu.Lock()
		service.repositories[id] = definition
		service.stateMu.Unlock()
		_ = service.save(definition)
		return err
	}
	definition.HeadOID = head
	definition.LastFetchedAt = time.Now().UTC().Format(time.RFC3339)
	definition.Status = "ready"
	service.stateMu.Lock()
	service.repositories[id] = definition
	service.stateMu.Unlock()
	return service.save(definition)
}

func (service *Service) Repositories(ctx context.Context) ([]Repository, error) {
	result := make([]Repository, 0, len(service.repositories))
	for _, definition := range OfficialRepositories() {
		service.stateMu.RLock()
		repository := service.repositories[definition.ID]
		service.stateMu.RUnlock()
		if service.git.exists(repository.ID) {
			repository.Status = "ready"
			if oid, err := service.git.resolve(ctx, repository.ID, "refs/heads/"+repository.DefaultBranch); err == nil {
				repository.HeadOID = oid
				if commit, commitErr := service.git.commit(ctx, repository.ID, oid); commitErr == nil {
					repository.HeadCommit = &commit
				}
			}
		}
		repository.RemoteURL = ""
		result = append(result, repository)
	}
	return result, nil
}

func (service *Service) Repository(ctx context.Context, id string) (Repository, error) {
	service.stateMu.RLock()
	repository, ok := service.repositories[id]
	service.stateMu.RUnlock()
	if !ok {
		return Repository{}, storage.ErrNotFound
	}
	if !service.git.exists(id) {
		return Repository{}, fmt.Errorf("repository %s is not mirrored yet", id)
	}
	branch, err := service.git.defaultBranch(ctx, id)
	if err == nil && branch != "" {
		repository.DefaultBranch = branch
	}
	repository.HeadOID, err = service.git.resolve(ctx, id, "refs/heads/"+repository.DefaultBranch)
	if err != nil {
		return Repository{}, err
	}
	repository.Status = "ready"
	repository.RemoteURL = ""
	return repository, nil
}

func (service *Service) Tree(ctx context.Context, id, ref, treePath string) (Tree, error) {
	repository, err := service.Repository(ctx, id)
	if err != nil {
		return Tree{}, err
	}
	if err := validateGitPath(treePath); err != nil {
		return Tree{}, err
	}
	if ref == "" {
		ref = "refs/heads/" + repository.DefaultBranch
	}
	oid, err := service.git.resolve(ctx, id, ref)
	if err != nil {
		return Tree{}, err
	}
	entries, err := service.git.tree(ctx, id, oid, treePath)
	if err != nil {
		return Tree{}, err
	}
	sortTreeEntries(entries)
	service.git.addLastCommits(ctx, id, oid, entries)
	commit, err := service.git.commit(ctx, id, oid)
	if err != nil {
		return Tree{}, err
	}
	result := Tree{Repository: repository, Ref: ref, Path: treePath, Commit: commit, Entries: entries}
	for _, entry := range entries {
		if entry.Type == "blob" && strings.EqualFold(entry.Name, "README.md") {
			readme, readErr := service.git.blob(ctx, id, oid, entry.Path)
			if readErr == nil && !readme.Binary {
				result.Readme = &readme
			}
			break
		}
	}
	if treePath == "" {
		allEntries, treeErr := service.git.recursiveTree(ctx, id, oid)
		if treeErr == nil {
			result.Languages = DetectLanguages(allEntries)
		}
	}
	return result, nil
}

func sortTreeEntries(entries []TreeEntry) {
	sort.SliceStable(entries, func(left, right int) bool {
		leftDirectory := entries[left].Type == "tree"
		rightDirectory := entries[right].Type == "tree"
		if leftDirectory != rightDirectory {
			return leftDirectory
		}
		leftName := strings.ToLower(entries[left].Name)
		rightName := strings.ToLower(entries[right].Name)
		if leftName == rightName {
			return entries[left].Name < entries[right].Name
		}
		return leftName < rightName
	})
}

func (service *Service) Blob(ctx context.Context, id, ref, blobPath string) (Blob, error) {
	if _, err := service.Repository(ctx, id); err != nil {
		return Blob{}, err
	}
	if err := validateGitPath(blobPath); err != nil || blobPath == "" {
		return Blob{}, fmt.Errorf("invalid blob path")
	}
	if ref == "" {
		service.stateMu.RLock()
		repository := service.repositories[id]
		service.stateMu.RUnlock()
		ref = "refs/heads/" + repository.DefaultBranch
	}
	oid, err := service.git.resolve(ctx, id, ref)
	if err != nil {
		return Blob{}, err
	}
	blob, err := service.git.blob(ctx, id, oid, blobPath)
	if err != nil {
		return Blob{}, err
	}
	if service.analyzer != nil && !blob.Binary && strings.EqualFold(filepath.Ext(blob.Path), ".wave") {
		highlight, analyzeErr := service.analyzer.Analyze(ctx, []byte(blob.Content))
		if analyzeErr == nil {
			blob.WaveHighlight = &highlight
		}
	}
	return blob, nil
}

func (service *Service) RawBlob(ctx context.Context, id, ref, blobPath string) ([]byte, string, error) {
	if _, err := service.Repository(ctx, id); err != nil {
		return nil, "", err
	}
	if err := validateGitPath(blobPath); err != nil || blobPath == "" {
		return nil, "", fmt.Errorf("invalid blob path")
	}
	if ref == "" {
		service.stateMu.RLock()
		repository := service.repositories[id]
		service.stateMu.RUnlock()
		ref = "refs/heads/" + repository.DefaultBranch
	}
	oid, err := service.git.resolve(ctx, id, ref)
	if err != nil {
		return nil, "", err
	}
	return service.git.rawBlob(ctx, id, oid, blobPath)
}

func (service *Service) Commits(ctx context.Context, id, ref, filePath string) ([]Commit, error) {
	repository, err := service.Repository(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := validateGitPath(filePath); err != nil {
		return nil, err
	}
	if ref == "" {
		ref = "refs/heads/" + repository.DefaultBranch
	}
	oid, err := service.git.resolve(ctx, id, ref)
	if err != nil {
		return nil, err
	}
	return service.git.commits(ctx, id, oid, filePath, 50)
}

func (service *Service) CommitDetail(ctx context.Context, id, oid string) (CommitDetail, error) {
	if _, err := service.Repository(ctx, id); err != nil {
		return CommitDetail{}, err
	}
	if !isCommitOID(oid) {
		return CommitDetail{}, fmt.Errorf("invalid commit oid")
	}
	resolved, err := service.git.resolve(ctx, id, oid)
	if err != nil {
		return CommitDetail{}, err
	}
	return service.git.commitDetail(ctx, id, resolved)
}

func isCommitOID(value string) bool {
	if len(value) < 7 || len(value) > 40 {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) {
			return false
		}
	}
	return true
}

func (service *Service) Refs(ctx context.Context, id string) (Refs, error) {
	if _, err := service.Repository(ctx, id); err != nil {
		return Refs{}, err
	}
	branches, err := service.git.refs(ctx, id, "refs/heads")
	if err != nil {
		return Refs{}, err
	}
	tags, err := service.git.refs(ctx, id, "refs/tags")
	if err != nil {
		return Refs{}, err
	}
	return Refs{Branches: branches, Tags: tags}, nil
}

func validateGitPath(value string) error {
	if value == "" {
		return nil
	}
	if strings.HasPrefix(value, "/") || path.Clean(value) != value || value == "." || strings.ContainsRune(value, '\x00') {
		return fmt.Errorf("invalid Git path")
	}
	return nil
}

func (service *Service) save(repository Repository) error {
	data, err := xml.Marshal(repository)
	if err != nil {
		return err
	}
	return service.database.Set(storage.Key("git", "repository", repository.ID), data)
}

func (service *Service) load(id string) (Repository, error) {
	data, err := service.database.Get(storage.Key("git", "repository", id))
	if err != nil {
		return Repository{}, err
	}
	var repository Repository
	if err := xml.Unmarshal(data, &repository); err != nil {
		return Repository{}, err
	}
	return repository, nil
}
