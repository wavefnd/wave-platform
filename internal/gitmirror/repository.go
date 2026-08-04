package gitmirror

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	maxBlobBytes    = 1 << 20
	maxRawBlobBytes = 16 << 20
	maxPatchBytes   = 2 << 20
)

type gitRepository struct {
	root string
}

func newGitRepository(root string) *gitRepository {
	return &gitRepository{root: root}
}

func (repository *gitRepository) path(id string) string {
	return filepath.Join(repository.root, id+".git")
}

func (repository *gitRepository) exists(id string) bool {
	info, err := os.Stat(repository.path(id))
	return err == nil && info.IsDir()
}

func (repository *gitRepository) sync(ctx context.Context, definition Repository) error {
	destination := repository.path(definition.ID)
	if repository.exists(definition.ID) {
		if _, err := repository.run(ctx, definition.ID, "remote", "set-url", "origin", definition.RemoteURL); err != nil {
			return err
		}
		if _, err := repository.run(ctx, definition.ID, "remote", "update", "--prune"); err != nil {
			return fmt.Errorf("update mirror %s: %w", definition.ID, err)
		}
		return nil
	}

	temporary, err := os.MkdirTemp(filepath.Dir(destination), "."+definition.ID+"-")
	if err != nil {
		return fmt.Errorf("create mirror staging directory: %w", err)
	}
	defer os.RemoveAll(temporary)

	staging := filepath.Join(temporary, definition.ID+".git")
	command := exec.CommandContext(ctx, "git", "clone", "--mirror", "--quiet", definition.RemoteURL, staging)
	command.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("clone mirror %s: %w: %s", definition.ID, err, strings.TrimSpace(string(output)))
	}
	if err := os.Rename(staging, destination); err != nil {
		return fmt.Errorf("publish mirror %s: %w", definition.ID, err)
	}
	return nil
}

func (repository *gitRepository) resolve(ctx context.Context, id, ref string) (string, error) {
	if ref == "" {
		ref = "HEAD"
	}
	output, err := repository.run(ctx, id, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (repository *gitRepository) defaultBranch(ctx context.Context, id string) (string, error) {
	output, err := repository.run(ctx, id, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimSpace(string(output)), "refs/heads/"), nil
}

func (repository *gitRepository) commit(ctx context.Context, id, oid string) (Commit, error) {
	output, err := repository.run(ctx, id, "show", "-s", "--format=%H%x1f%h%x1f%an%x1f%aI%x1f%s", oid)
	if err != nil {
		return Commit{}, err
	}
	return parseCommit(strings.TrimSpace(string(output)))
}

func (repository *gitRepository) commits(ctx context.Context, id, oid, path string, limit int) ([]Commit, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	arguments := []string{"log", "--max-count=" + strconv.Itoa(limit), "--format=%H%x1f%h%x1f%an%x1f%aI%x1f%s%x1e", oid}
	if path != "" {
		arguments = append(arguments, "--", path)
	}
	output, err := repository.run(ctx, id, arguments...)
	if err != nil {
		return nil, err
	}
	result := make([]Commit, 0)
	for _, record := range strings.Split(string(output), "\x1e") {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}
		commit, err := parseCommit(record)
		if err != nil {
			return nil, err
		}
		result = append(result, commit)
	}
	return result, nil
}

func (repository *gitRepository) commitDetail(ctx context.Context, id, oid string) (CommitDetail, error) {
	commit, err := repository.commit(ctx, id, oid)
	if err != nil {
		return CommitDetail{}, err
	}

	metadata, err := repository.run(ctx, id, "show", "-s", "--format=%P%x1f%B", oid)
	if err != nil {
		return CommitDetail{}, err
	}
	fields := strings.SplitN(strings.TrimSpace(string(metadata)), "\x1f", 2)
	parents := []string{}
	body := ""
	if len(fields) > 0 && strings.TrimSpace(fields[0]) != "" {
		parents = strings.Fields(fields[0])
	}
	if len(fields) == 2 {
		body = strings.TrimSpace(fields[1])
		if body == commit.Subject {
			body = ""
		} else if strings.HasPrefix(body, commit.Subject+"\n") {
			body = strings.TrimSpace(strings.TrimPrefix(body, commit.Subject))
		}
	}

	changed, err := repository.run(ctx, id, "diff-tree", "--root", "--no-commit-id", "--name-status", "-r", "-z", "--find-renames", oid)
	if err != nil {
		return CommitDetail{}, err
	}
	files, err := parseChangedFiles(changed)
	if err != nil {
		return CommitDetail{}, err
	}

	patch, err := repository.run(ctx, id, "show", "--format=", "--no-color", "--no-ext-diff", "--find-renames", "--unified=3", oid)
	if err != nil {
		return CommitDetail{}, err
	}
	truncated := len(patch) > maxPatchBytes
	if truncated {
		patch = patch[:maxPatchBytes]
	}

	return CommitDetail{
		Commit: commit, Body: body, Parents: parents, Files: files,
		Patch: string(patch), PatchTruncated: truncated,
	}, nil
}

func parseChangedFiles(value []byte) ([]ChangedFile, error) {
	fields := bytes.Split(value, []byte{0})
	files := make([]ChangedFile, 0)
	for index := 0; index < len(fields); {
		status := strings.TrimSpace(string(fields[index]))
		index++
		if status == "" {
			continue
		}
		if index >= len(fields) || len(status) > 4 {
			return nil, fmt.Errorf("parse changed file status")
		}
		firstPath := string(fields[index])
		index++
		file := ChangedFile{Status: status[:1], Path: firstPath}
		if status[0] == 'R' || status[0] == 'C' {
			if index >= len(fields) {
				return nil, fmt.Errorf("parse renamed file")
			}
			file.OldPath = firstPath
			file.Path = string(fields[index])
			index++
		}
		files = append(files, file)
	}
	return files, nil
}

func parseCommit(value string) (Commit, error) {
	fields := strings.SplitN(value, "\x1f", 5)
	if len(fields) != 5 {
		return Commit{}, fmt.Errorf("parse git commit record")
	}
	return Commit{OID: fields[0], ShortOID: fields[1], Author: fields[2], AuthoredAt: fields[3], Subject: fields[4]}, nil
}

func (repository *gitRepository) tree(ctx context.Context, id, oid, treePath string) ([]TreeEntry, error) {
	treeish := oid
	if treePath != "" {
		treeish += ":" + treePath
	}
	output, err := repository.run(ctx, id, "ls-tree", "-z", "-l", treeish)
	if err != nil {
		return nil, err
	}
	entries := make([]TreeEntry, 0)
	for _, record := range bytes.Split(output, []byte{0}) {
		if len(record) == 0 {
			continue
		}
		parts := bytes.SplitN(record, []byte{'\t'}, 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("parse git tree entry")
		}
		metadata := strings.Fields(string(parts[0]))
		if len(metadata) != 4 {
			return nil, fmt.Errorf("parse git tree metadata")
		}
		name := string(parts[1])
		entryPath := name
		if treePath != "" {
			entryPath = treePath + "/" + name
		}
		entry := TreeEntry{Name: name, Path: entryPath, Type: metadata[1], OID: metadata[2]}
		if metadata[3] != "-" {
			entry.Size, _ = strconv.ParseInt(metadata[3], 10, 64)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (repository *gitRepository) addLastCommits(ctx context.Context, id, oid string, entries []TreeEntry) {
	var wait sync.WaitGroup
	workers := make(chan struct{}, 8)
	for index := range entries {
		index := index
		wait.Add(1)
		go func() {
			defer wait.Done()
			workers <- struct{}{}
			defer func() { <-workers }()
			commits, err := repository.commits(ctx, id, oid, entries[index].Path, 1)
			if err == nil && len(commits) == 1 {
				entries[index].LastCommit = &commits[0]
			}
		}()
	}
	wait.Wait()
}

func (repository *gitRepository) recursiveTree(ctx context.Context, id, oid string) ([]TreeEntry, error) {
	output, err := repository.run(ctx, id, "ls-tree", "-r", "-z", "-l", oid)
	if err != nil {
		return nil, err
	}
	entries := make([]TreeEntry, 0)
	for _, record := range bytes.Split(output, []byte{0}) {
		parts := bytes.SplitN(record, []byte{'\t'}, 2)
		if len(parts) != 2 {
			continue
		}
		metadata := strings.Fields(string(parts[0]))
		if len(metadata) != 4 || metadata[1] != "blob" {
			continue
		}
		size, _ := strconv.ParseInt(metadata[3], 10, 64)
		filePath := string(parts[1])
		entries = append(entries, TreeEntry{Path: filePath, Name: filepath.Base(filePath), Type: "blob", OID: metadata[2], Size: size, Generated: isGenerated(filePath)})
	}
	return entries, nil
}

func isGenerated(path string) bool {
	lower := strings.ToLower(path)
	return strings.Contains(lower, "/vendor/") || strings.Contains(lower, "/generated/") || strings.HasSuffix(lower, ".min.js")
}

func (repository *gitRepository) blob(ctx context.Context, id, oid, path string) (Blob, error) {
	entryOutput, err := repository.run(ctx, id, "ls-tree", "-l", oid, "--", path)
	if err != nil {
		return Blob{}, err
	}
	metadata := strings.Fields(strings.SplitN(string(entryOutput), "\t", 2)[0])
	if len(metadata) != 4 || metadata[1] != "blob" {
		return Blob{}, errors.New("path is not a blob")
	}
	size, _ := strconv.ParseInt(metadata[3], 10, 64)
	output, err := repository.run(ctx, id, "show", oid+":"+path)
	if err != nil {
		return Blob{}, err
	}
	truncated := len(output) > maxBlobBytes
	if truncated {
		output = output[:maxBlobBytes]
	}
	binary := bytes.IndexByte(output, 0) >= 0 || !utf8.Valid(output) || hasInvalidXMLControl(output)
	content := ""
	if !binary {
		content = string(output)
	}
	return Blob{Path: path, OID: metadata[2], Size: size, Binary: binary, Truncated: truncated, Content: content}, nil
}

func (repository *gitRepository) rawBlob(ctx context.Context, id, oid, path string) ([]byte, string, error) {
	entryOutput, err := repository.run(ctx, id, "ls-tree", "-l", oid, "--", path)
	if err != nil {
		return nil, "", err
	}
	metadata := strings.Fields(strings.SplitN(string(entryOutput), "\t", 2)[0])
	if len(metadata) != 4 || metadata[1] != "blob" {
		return nil, "", errors.New("path is not a blob")
	}
	size, err := strconv.ParseInt(metadata[3], 10, 64)
	if err != nil || size > maxRawBlobBytes {
		return nil, "", errors.New("raw blob exceeds size limit")
	}
	output, err := repository.run(ctx, id, "show", oid+":"+path)
	if err != nil {
		return nil, "", err
	}
	return output, metadata[2], nil
}

func hasInvalidXMLControl(value []byte) bool {
	for _, character := range value {
		if character < 0x20 && character != '\n' && character != '\r' && character != '\t' {
			return true
		}
	}
	return false
}

func (repository *gitRepository) refs(ctx context.Context, id, prefix string) ([]Ref, error) {
	output, err := repository.run(ctx, id, "for-each-ref", "--sort=-creatordate", "--format=%(refname:short)%00%(objectname)%00%(creatordate:iso-strict)", prefix)
	if err != nil {
		return nil, err
	}
	result := make([]Ref, 0)
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Split(line, "\x00")
		if len(fields) == 3 {
			result = append(result, Ref{Name: fields[0], OID: fields[1], UpdatedAt: fields[2]})
		}
	}
	return result, nil
}

func (repository *gitRepository) run(ctx context.Context, id string, arguments ...string) ([]byte, error) {
	if !repository.exists(id) {
		return nil, os.ErrNotExist
	}
	commandArguments := append([]string{"--git-dir=" + repository.path(id), "--no-pager"}, arguments...)
	command := exec.CommandContext(ctx, "git", commandArguments...)
	command.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(arguments, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func withTimeout(parent context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := parent.Deadline(); ok {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, duration)
}
