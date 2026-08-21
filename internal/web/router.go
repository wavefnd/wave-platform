package web

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	admindomain "github.com/wavefnd/wave-platform/internal/admin"
	blogdomain "github.com/wavefnd/wave-platform/internal/blog"
	communitydomain "github.com/wavefnd/wave-platform/internal/community"
	documentdomain "github.com/wavefnd/wave-platform/internal/document"
	"github.com/wavefnd/wave-platform/internal/gitmirror"
	mediadomain "github.com/wavefnd/wave-platform/internal/media"
	patchdomain "github.com/wavefnd/wave-platform/internal/patcharchive"
	"github.com/wavefnd/wave-platform/internal/platformstats"
	questiondomain "github.com/wavefnd/wave-platform/internal/question"
	"github.com/wavefnd/wave-platform/internal/sponsor"
	"github.com/wavefnd/wave-platform/internal/web/handler"
	webhookdomain "github.com/wavefnd/wave-platform/internal/webhook"
)

func NewRouter(
	environment string,
	frontendPath string,
	publicURL string,
	version string,
	modules []handler.ModuleStatus,
	databaseCheck func() error,
	blogService *blogdomain.Service,
	documentRepository *documentdomain.Repository,
	communityRepository *communitydomain.Repository,
	communityService *communitydomain.Service,
	questionRepository *questiondomain.Repository,
	questionService *questiondomain.Service,
	sourceService *gitmirror.Service,
	statsService *platformstats.Service,
	authHandler *handler.AuthHandler,
	mailboxHandler *handler.MailboxHandler,
	adminService *admindomain.Service,
	mediaService *mediadomain.Service,
	webhookService *webhookdomain.Service,
	patchService *patchdomain.Service,
) http.Handler {
	mux := http.NewServeMux()

	platformHandler := handler.PlatformHandler{
		Environment: environment,
		Version:     version,
	}
	modulesHandler := handler.ModulesHandler{Modules: modules, Auth: authHandler}
	healthHandler := handler.HealthHandler{DatabaseCheck: databaseCheck}
	releasesHandler := handler.ReleasesHandler{Service: blogService}
	blogHandler := handler.BlogHandler{Service: blogService, Auth: authHandler}
	documentsHandler := handler.DocumentsHandler{Repository: documentRepository}
	communityHandler := handler.CommunityHandler{Repository: communityRepository, Service: communityService, Auth: authHandler}
	questionsHandler := handler.QuestionsHandler{Repository: questionRepository, Service: questionService, Auth: authHandler}
	sourceHandler := handler.SourceHandler{Service: sourceService}
	statsHandler := handler.StatsHandler{Service: statsService}
	adminHandler := handler.AdministrationHandler{Service: adminService, Auth: authHandler}
	sponsorsHandler := handler.SponsorsHandler{Service: sponsor.NewService()}
	mediaHandler := handler.MediaHandler{Service: mediaService, Auth: authHandler}
	webhookHandler := handler.WebhookHandler{Service: webhookService, Auth: authHandler}
	patchesHandler := handler.PatchesHandler{Service: patchService}
	usersHandler := handler.UsersHandler{Community: communityRepository, Questions: questionRepository, Auth: authHandler}
	seoHandler := NewSEOHandler(publicURL, documentRepository, blogService, communityRepository, questionRepository)

	mux.HandleFunc("GET /api/v1/platform", platformHandler.Status)
	mux.HandleFunc("GET /api/v1/modules", modulesHandler.Status)
	mux.HandleFunc("GET /api/v1/releases", releasesHandler.List)
	mux.HandleFunc("GET /api/v1/releases/{slug}", releasesHandler.Get)
	mux.HandleFunc("GET /api/v1/blog/posts", blogHandler.List)
	mux.HandleFunc("GET /api/v1/blog/posts/{slug}", blogHandler.Get)
	mux.HandleFunc("GET /api/v1/sponsors", sponsorsHandler.List)
	mux.HandleFunc("GET /api/v1/documents", documentsHandler.List)
	mux.HandleFunc("GET /api/v1/documents/{path...}", documentsHandler.Get)
	mux.HandleFunc("GET /api/v1/community/spaces", communityHandler.Spaces)
	mux.HandleFunc("GET /api/v1/community/threads", communityHandler.Threads)
	mux.HandleFunc("POST /api/v1/community/threads", communityHandler.CreatePost)
	mux.HandleFunc("GET /api/v1/community/threads/{thread}", communityHandler.Thread)
	mux.HandleFunc("POST /api/v1/community/threads/{thread}/comments", communityHandler.CreateReply)
	mux.HandleFunc("POST /api/v1/community/threads/{thread}/vote", communityHandler.Vote)
	mux.HandleFunc("POST /api/v1/community/threads/{thread}/subscription", communityHandler.Subscribe)
	mux.HandleFunc("GET /api/v1/questions", questionsHandler.List)
	mux.HandleFunc("POST /api/v1/questions", questionsHandler.Create)
	mux.HandleFunc("GET /api/v1/questions/{question}", questionsHandler.Get)
	mux.HandleFunc("POST /api/v1/questions/{question}/answers", questionsHandler.Answer)
	mux.HandleFunc("POST /api/v1/questions/{question}/vote", questionsHandler.Vote)
	mux.HandleFunc("POST /api/v1/questions/{question}/accept", questionsHandler.Accept)
	mux.HandleFunc("GET /api/v1/source/repositories", sourceHandler.Repositories)
	mux.HandleFunc("GET /api/v1/source/repositories/{repository}/tree", sourceHandler.Tree)
	mux.HandleFunc("GET /api/v1/source/repositories/{repository}/blob", sourceHandler.Blob)
	mux.HandleFunc("GET /api/v1/source/repositories/{repository}/raw", sourceHandler.RawBlob)
	mux.HandleFunc("GET /api/v1/source/repositories/{repository}/commits", sourceHandler.Commits)
	mux.HandleFunc("GET /api/v1/source/repositories/{repository}/commits/{oid}", sourceHandler.CommitDetail)
	mux.HandleFunc("GET /api/v1/source/repositories/{repository}/refs", sourceHandler.Refs)
	mux.HandleFunc("GET /api/v1/patches", patchesHandler.List)
	mux.HandleFunc("GET /api/v1/patches/{patch}", patchesHandler.Get)
	mux.HandleFunc("GET /api/v1/platform/stats", statsHandler.Get)
	mux.HandleFunc("GET /api/v1/platform/preferences", adminHandler.PlatformPreferences)
	mux.HandleFunc("GET /api/v1/users", usersHandler.Directory)
	mux.HandleFunc("GET /api/v1/users/by-id/{account}", usersHandler.ProfileByID)
	mux.HandleFunc("GET /api/v1/users/{user}", usersHandler.Profile)
	mux.HandleFunc("POST /api/v1/users/me/profile", usersHandler.UpdateProfile)
	mux.HandleFunc("POST /api/v1/users/me/address", usersHandler.UpdateAddress)
	mux.HandleFunc("GET /api/v1/admin", adminHandler.Snapshot)
	mux.HandleFunc("POST /api/v1/admin/accounts/{account}/status", adminHandler.AccountStatus)
	mux.HandleFunc("POST /api/v1/admin/accounts/{account}/role", adminHandler.AccountRole)
	mux.HandleFunc("POST /api/v1/admin/settings/lunastev-time-zone", adminHandler.LunaStevTimeZone)
	mux.HandleFunc("GET /api/v1/admin/blog/posts", blogHandler.AdminList)
	mux.HandleFunc("GET /api/v1/admin/blog/posts/{slug}", blogHandler.AdminGet)
	mux.HandleFunc("POST /api/v1/admin/blog/posts", blogHandler.Save)
	mux.HandleFunc("GET /api/v1/admin/webhooks", webhookHandler.List)
	mux.HandleFunc("POST /api/v1/admin/webhooks", webhookHandler.Save)
	mux.HandleFunc("DELETE /api/v1/admin/webhooks/{webhook}", webhookHandler.Delete)
	mux.HandleFunc("POST /api/v1/admin/webhooks/{webhook}/test", webhookHandler.Test)
	mux.HandleFunc("GET /api/v1/webhooks", webhookHandler.UserList)
	mux.HandleFunc("POST /api/v1/webhooks", webhookHandler.UserSave)
	mux.HandleFunc("DELETE /api/v1/webhooks/{webhook}", webhookHandler.UserDelete)
	mux.HandleFunc("POST /api/v1/webhooks/{webhook}/test", webhookHandler.UserTest)
	mux.HandleFunc("POST /api/v1/media/lunastev/images", mediaHandler.UploadLunaStevImage)
	mux.HandleFunc("GET /media/lunastev/{image}", mediaHandler.LunaStevImage)
	if authHandler != nil {
		mux.HandleFunc("GET /api/v1/auth/config", authHandler.Config)
		mux.HandleFunc("GET /api/v1/auth/registration-address", authHandler.RegistrationAddress)
		mux.HandleFunc("POST /api/v1/auth/register/begin", authHandler.BeginRegistration)
		mux.HandleFunc("POST /api/v1/auth/register/finish", authHandler.FinishRegistration)
		mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
		mux.HandleFunc("POST /api/v1/auth/recovery/request", authHandler.RequestRecovery)
		mux.HandleFunc("POST /api/v1/auth/recovery/enrollment", authHandler.RecoveryEnrollment)
		mux.HandleFunc("POST /api/v1/auth/recovery/finish", authHandler.FinishRecovery)
		mux.HandleFunc("GET /api/v1/auth/security", authHandler.Security)
		mux.HandleFunc("POST /api/v1/auth/security/totp/begin", authHandler.BeginRotation)
		mux.HandleFunc("POST /api/v1/auth/security/totp/finish", authHandler.FinishRotation)
		mux.HandleFunc("POST /api/v1/auth/security/recovery-email", authHandler.ChangeRecoveryEmail)
		mux.HandleFunc("POST /api/v1/auth/recovery-email/verify", authHandler.VerifyRecoveryEmail)
		mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)
		mux.HandleFunc("GET /api/v1/auth/session", authHandler.Current)
	}
	if mailboxHandler != nil {
		mux.HandleFunc("GET /api/v1/mailbox", mailboxHandler.List)
		mux.HandleFunc("POST /api/v1/mailbox/messages", mailboxHandler.Send)
		mux.HandleFunc("GET /api/v1/mailbox/messages/{entry}", mailboxHandler.Message)
		mux.HandleFunc("POST /api/v1/mailbox/messages/{entry}/action", mailboxHandler.Action)
		mux.HandleFunc("GET /api/v1/admin/mailbox", mailboxHandler.ManagementList)
		mux.HandleFunc("GET /api/v1/admin/mailbox/messages/{entry}", mailboxHandler.ManagementMessage)
		mux.HandleFunc("POST /api/v1/admin/mailbox/messages/{entry}/action", mailboxHandler.ManagementAction)
	}
	mux.HandleFunc("GET /health", healthHandler.Ready)
	mux.HandleFunc("GET /health/live", healthHandler.Live)
	mux.HandleFunc("GET /health/ready", healthHandler.Ready)
	mux.HandleFunc("GET /robots.txt", seoHandler.Robots)
	mux.HandleFunc("GET /sitemap.xml", seoHandler.Sitemap)
	mux.HandleFunc("GET /community/announcements/{slug}", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/blog/"+request.PathValue("slug"), http.StatusPermanentRedirect)
	})

	mux.Handle("/", frontendHandler(frontendPath, seoHandler))

	return mux
}

func frontendHandler(root string, seo SEOHandler) http.Handler {
	fileServer := http.FileServer(http.Dir(root))
	indexPath := filepath.Join(root, "index.html")
	indexDocument, _ := os.ReadFile(indexPath)
	const seoStart = "<!-- wave:seo:start -->"
	const seoEnd = "<!-- wave:seo:end -->"
	renderIndex := func(writer http.ResponseWriter, request *http.Request) {
		document := indexDocument
		start := bytes.Index(document, []byte(seoStart))
		end := bytes.Index(document, []byte(seoEnd))
		if start >= 0 && end > start {
			end += len(seoEnd)
			replacement := []byte(seoStart + "\n    " + seo.HTMLMetadata(request) + "\n    " + seoEnd)
			document = append(append(append([]byte{}, document[:start]...), replacement...), document[end:]...)
		}
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.Header().Set("Cache-Control", "no-cache")
		_, _ = writer.Write(document)
	}

	return http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if privateFrontendPath(request.URL.Path) {
			writer.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
		}
		if request.URL.Path == "/" {
			renderIndex(writer, request)
			return
		}

		requestedPath := filepath.Join(
			root,
			filepath.Clean(request.URL.Path),
		)

		if info, err := os.Stat(requestedPath); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(writer, request)
			return
		}

		if strings.HasPrefix(request.URL.Path, "/api/") {
			http.NotFound(writer, request)
			return
		}

		renderIndex(writer, request)
	})
}

func privateFrontendPath(value string) bool {
	for _, prefix := range []string{"/admin", "/account", "/login", "/register", "/mail"} {
		if value == prefix || strings.HasPrefix(value, prefix+"/") {
			return true
		}
	}
	return false
}
