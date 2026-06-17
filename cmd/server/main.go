package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bryansmee/homeprojects/internal/api"
	"github.com/bryansmee/homeprojects/internal/auth"
	"github.com/bryansmee/homeprojects/internal/authz"
	"github.com/bryansmee/homeprojects/internal/config"
	"github.com/bryansmee/homeprojects/internal/db"
	"github.com/bryansmee/homeprojects/internal/extensions"
	"github.com/bryansmee/homeprojects/internal/extensions/printing"
	"github.com/bryansmee/homeprojects/internal/models"
	"github.com/bryansmee/homeprojects/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	gdb, err := db.Open(cfg.DB.Driver, cfg.DB.DSN)
	if err != nil {
		return err
	}

	registry := extensions.NewRegistry(printing.New())
	if err := db.Migrate(gdb, append(models.AllModels(), registry.Models()...)...); err != nil {
		return err
	}

	azEngine, err := authz.New(ctx)
	if err != nil {
		return err
	}

	sessions := auth.NewSessionManager(
		cfg.Session.Secret, cfg.Session.TTL, cfg.Session.CookieName,
		cfg.Session.CookieSecure, cfg.Session.CookieDomain,
	)

	authn, err := buildAuthenticator(ctx, cfg)
	if err != nil {
		return err
	}

	srv := api.NewServer(cfg, store.New(gdb), azEngine, sessions, authn, registry)
	return serve(cfg.HTTP.Addr, srv.Routes())
}

func buildAuthenticator(ctx context.Context, cfg *config.Config) (*auth.Authenticator, error) {
	if !cfg.OIDCEnabled() {
		log.Printf("OIDC disabled: dev-login is enabled at POST /api/auth/dev-login")
		return nil, nil
	}
	return auth.NewAuthenticator(
		ctx, cfg.OIDC.Issuer, cfg.OIDC.ClientID, cfg.OIDC.ClientSecret,
		cfg.OIDC.RedirectURL, cfg.OIDC.Scopes,
	)
}

func serve(addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", addr)
		errCh <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-stop:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Printf("shutting down")
		return srv.Shutdown(shutdownCtx)
	}
}
