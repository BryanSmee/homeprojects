package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/bryansmee/homeprojects/internal/auth"
	"github.com/bryansmee/homeprojects/internal/authz"
	"github.com/bryansmee/homeprojects/internal/config"
	"github.com/bryansmee/homeprojects/internal/extensions"
	"github.com/bryansmee/homeprojects/internal/store"
)

type Server struct {
	cfg        *config.Config
	store      *store.Store
	authz      *authz.Engine
	sessions   *auth.SessionManager
	auth       *auth.Authenticator // nil in dev mode
	extensions *extensions.Registry
}

func NewServer(
	cfg *config.Config,
	st *store.Store,
	az *authz.Engine,
	sessions *auth.SessionManager,
	authn *auth.Authenticator,
	reg *extensions.Registry,
) *Server {
	return &Server{cfg: cfg, store: st, authz: az, sessions: sessions, auth: authn, extensions: reg}
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api", func(api chi.Router) {
		api.Use(s.sessionMiddleware)
		s.mountAuthRoutes(api)
		s.mountProjectRoutes(api)
		s.mountExtensionRoutes(api)
	})
	return r
}

func (s *Server) mountExtensionRoutes(api chi.Router) {
	deps := extensions.Deps{DB: s.store.DB(), Authz: s}
	for _, ext := range s.extensions.All() {
		ext := ext
		api.Route("/ext/"+ext.Name(), func(er chi.Router) {
			ext.Mount(er, deps)
		})
	}
}
