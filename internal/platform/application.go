package platform

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	admindomain "github.com/wavefnd/wave-platform/internal/admin"
	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/auth"
	"github.com/wavefnd/wave-platform/internal/community"
	"github.com/wavefnd/wave-platform/internal/config"
	"github.com/wavefnd/wave-platform/internal/document"
	"github.com/wavefnd/wave-platform/internal/gitmirror"
	"github.com/wavefnd/wave-platform/internal/identity"
	"github.com/wavefnd/wave-platform/internal/mailruntime"
	mediadomain "github.com/wavefnd/wave-platform/internal/media"
	"github.com/wavefnd/wave-platform/internal/mediapolicy"
	"github.com/wavefnd/wave-platform/internal/platformstats"
	questiondomain "github.com/wavefnd/wave-platform/internal/question"
	releasedomain "github.com/wavefnd/wave-platform/internal/release"
	"github.com/wavefnd/wave-platform/internal/sourceanalysis"
	"github.com/wavefnd/wave-platform/internal/storage"
	"github.com/wavefnd/wave-platform/internal/waveruntime"
	"github.com/wavefnd/wave-platform/internal/web"
	"github.com/wavefnd/wave-platform/internal/web/handler"
)

const Version = "0.1.0"

type Application struct {
	Config         *config.Config
	Database       *storage.Database
	Server         *http.Server
	GitMirror      *gitmirror.Service
	SourceAnalyzer sourceanalysis.Analyzer
	MediaPolicy    *waveruntime.NativeMediaPolicy
	Identity       *identity.Service
	MailRuntime    *mailruntime.Service
}

func New(configPath string) (*Application, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	database, err := storage.Open(cfg.Storage.Root)
	if err != nil {
		return nil, err
	}
	identityService, err := identity.NewServiceWithTOTP(database, cfg.Identity.MailDomain, cfg.Identity.RegistrationOpen,
		time.Duration(cfg.Identity.SessionHours)*time.Hour, cfg.Identity.AuthEncryptionKey, cfg.Identity.TOTPIssuer, cfg.Identity.PublicURL)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("initialize TOTP authentication: %w", err)
	}
	adminAccount, created, err := identityService.BootstrapTOTPAdmin(cfg.Identity.AdminDisplayName, cfg.Identity.AdminUsername,
		cfg.Identity.AdminRecoveryEmail, cfg.Identity.AdminTOTPSecret)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("bootstrap administrator: %w", err)
	}
	if created {
		log.Printf("Wave administrator created: %s", adminAccount.Email)
	}
	if _, err := community.SeedLanguageReleases(database); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("seed language releases: %w", err)
	}
	if err := community.SeedSpaces(database); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("seed community spaces: %w", err)
	}
	if count, err := document.SeedOfficial(database); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("seed official documents: %w", err)
	} else if count > 0 {
		log.Printf("Published %d official document translations", count)
	}
	removedCommunityEntries, err := community.CleanupMailboxProjections(database)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("clean community mailbox projections: %w", err)
	}
	if removedCommunityEntries > 0 {
		log.Printf("Removed %d legacy community mailbox entries", removedCommunityEntries)
	}
	releaseRepository := releasedomain.NewRepository(database)
	documentRepository := document.NewRepository(database)
	communityRepository := community.NewRepository(database)
	communityService := community.NewService(database, cfg.Identity.MailDomain)
	questionRepository := questiondomain.NewRepository(database)
	questionService := questiondomain.NewService(database, cfg.Identity.MailDomain)
	var sourceAnalyzer sourceanalysis.Analyzer
	var mediaPlanner mediapolicy.Planner
	var nativeMediaPolicy *waveruntime.NativeMediaPolicy
	if cfg.Wave.Enabled {
		modulePath := filepath.Join(cfg.Wave.Modules, "libwave-source-analyzer.so")
		loadedAnalyzer, loadErr := waveruntime.OpenSourceAnalyzer(modulePath)
		if loadErr != nil {
			log.Printf("Wave source analyzer unavailable: %v", loadErr)
		} else {
			sourceAnalyzer = loadedAnalyzer
			log.Printf("Wave source analyzer ready: %s", modulePath)
		}
		mediaPolicyPath := filepath.Join(cfg.Wave.Modules, "libwave-media-policy.so")
		loadedMediaPolicy, policyErr := waveruntime.OpenMediaPolicy(mediaPolicyPath)
		if policyErr != nil {
			log.Printf("Wave media policy unavailable: %v", policyErr)
		} else {
			nativeMediaPolicy = loadedMediaPolicy
			mediaPlanner = loadedMediaPolicy
			log.Printf("Wave media policy ready: %s", mediaPolicyPath)
		}
	}
	mediaService, err := mediadomain.NewService(filepath.Join(cfg.Storage.Root, "blobs", "lunastev", "images"), mediaPlanner, audit.NewRepository(database))
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("initialize LunaStev media: %w", err)
	}
	gitMirrorService, err := gitmirror.NewService(database, sourceAnalyzer)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("initialize Git mirror: %w", err)
	}
	statsService := platformstats.NewService(database, gitMirrorService)
	mailRuntime, err := mailruntime.New(cfg.Mail, database, identityService)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("initialize mail runtime: %w", err)
	}
	authHandler := &handler.AuthHandler{Service: identityService, MailDomain: cfg.Identity.MailDomain,
		RegistrationOpen: cfg.Identity.RegistrationOpen, SecureCookies: cfg.Identity.SecureCookies,
		Challenge: auth.TurnstileVerifier{SiteKey: cfg.Identity.TurnstileSiteKey, Secret: cfg.Identity.TurnstileSecret}}
	mailboxHandler := &handler.MailboxHandler{Auth: *authHandler}
	adminService := admindomain.NewService(database, gitMirrorService, cfg.Identity.RegistrationOpen,
		strings.TrimSpace(cfg.Identity.TurnstileSiteKey) != "")

	moduleStatuses := cfg.Modules.Statuses()
	modules := make([]handler.ModuleStatus, 0, len(moduleStatuses)+1)
	for _, module := range moduleStatuses {
		status := "disabled"
		if module.Enabled {
			status = "foundation"
		}

		modules = append(modules, handler.ModuleStatus{
			Name:    module.Name,
			Enabled: module.Enabled,
			Status:  status,
		})
	}

	modules = append(modules, handler.ModuleStatus{
		Name:    "wave-runtime",
		Enabled: cfg.Wave.Enabled,
		Status:  map[bool]string{true: "foundation", false: "disabled"}[cfg.Wave.Enabled],
	})

	router := web.NewRouter(
		cfg.Server.Environment,
		cfg.Frontend.Distribution,
		cfg.Identity.PublicURL,
		Version,
		modules,
		database.Health,
		releaseRepository,
		documentRepository,
		communityRepository,
		communityService,
		questionRepository,
		questionService,
		gitMirrorService,
		statsService,
		authHandler,
		mailboxHandler,
		adminService,
		mediaService,
	)

	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	return &Application{
		Config:         cfg,
		Database:       database,
		Server:         server,
		GitMirror:      gitMirrorService,
		SourceAnalyzer: sourceAnalyzer,
		MediaPolicy:    nativeMediaPolicy,
		Identity:       identityService,
		MailRuntime:    mailRuntime,
	}, nil
}

func (application *Application) Run(ctx context.Context) error {
	log.Printf(
		"Wave Platform listening on %s",
		application.Config.Server.Address,
	)

	runtimeContext, stopRuntime := context.WithCancel(ctx)
	var gitMirrorDone chan struct{}
	if application.Config.Modules.GitMirror.Enabled && application.GitMirror != nil {
		gitMirrorDone = make(chan struct{})
		go func() {
			defer close(gitMirrorDone)
			application.GitMirror.Run(runtimeContext)
		}()
	}
	defer func() {
		stopRuntime()
		if gitMirrorDone != nil {
			<-gitMirrorDone
		}
	}()

	type runtimeResult struct {
		name string
		err  error
	}
	result := make(chan runtimeResult, 2)
	go func() {
		result <- runtimeResult{name: "HTTP", err: application.Server.ListenAndServe()}
	}()
	go func() {
		result <- runtimeResult{name: "mail", err: application.MailRuntime.Run(runtimeContext)}
	}()

	select {
	case current := <-result:
		if current.err != nil && current.err != http.ErrServerClosed {
			return fmt.Errorf("serve %s: %w", current.name, current.err)
		}
		return nil
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := application.Server.Shutdown(shutdownContext); err != nil {
			return fmt.Errorf("shutdown HTTP: %w", err)
		}

		stopRuntime()
	}

	return nil
}

func (application *Application) Close() error {
	var closeErrors []error
	if application.SourceAnalyzer != nil {
		closeErrors = append(closeErrors, application.SourceAnalyzer.Close())
	}
	if application.MediaPolicy != nil {
		closeErrors = append(closeErrors, application.MediaPolicy.Close())
	}
	if application.Database != nil {
		closeErrors = append(closeErrors, application.Database.Close())
	}
	return errors.Join(closeErrors...)
}
