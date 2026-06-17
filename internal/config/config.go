// Package config loads runtime configuration from environment variables, with
// dev-friendly defaults so the server boots with zero config against SQLite.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTP    HTTPConfig
	DB      DBConfig
	OIDC    OIDCConfig
	Session SessionConfig
}

type HTTPConfig struct {
	Addr            string
	AllowedOrigins  []string
	PublicBaseURL   string // public URL of this API, used to build the OIDC redirect URL
	FrontendBaseURL string // where to send the browser after login
}

// DBConfig selects a registered DB driver (see internal/db).
type DBConfig struct {
	Driver string
	DSN    string
}

// OIDCConfig configures SSO (e.g. Pocket ID). An empty Issuer enables dev-login.
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type SessionConfig struct {
	Secret       string
	TTL          time.Duration
	CookieName   string
	CookieSecure bool
	CookieDomain string
}

func Load() (*Config, error) {
	c := &Config{
		HTTP: HTTPConfig{
			Addr:            env("HP_ADDR", ":8080"),
			AllowedOrigins:  splitList(env("HP_ALLOWED_ORIGINS", "http://localhost:3000")),
			PublicBaseURL:   env("HP_PUBLIC_BASE_URL", "http://localhost:8080"),
			FrontendBaseURL: env("HP_FRONTEND_BASE_URL", "http://localhost:3000"),
		},
		DB: DBConfig{
			Driver: env("HP_DB_DRIVER", "sqlite"),
			DSN:    env("HP_DB_DSN", "homeprojects.db"),
		},
		OIDC: OIDCConfig{
			Issuer:       env("HP_OIDC_ISSUER", ""),
			ClientID:     env("HP_OIDC_CLIENT_ID", ""),
			ClientSecret: env("HP_OIDC_CLIENT_SECRET", ""),
			RedirectURL:  env("HP_OIDC_REDIRECT_URL", ""),
			Scopes:       splitList(env("HP_OIDC_SCOPES", "openid,profile,email")),
		},
		Session: SessionConfig{
			Secret:       env("HP_SESSION_SECRET", "dev-insecure-change-me"),
			TTL:          envDuration("HP_SESSION_TTL", 7*24*time.Hour),
			CookieName:   env("HP_COOKIE_NAME", "hp_session"),
			CookieSecure: envBool("HP_COOKIE_SECURE", false),
			CookieDomain: env("HP_COOKIE_DOMAIN", ""),
		},
	}

	if c.OIDC.RedirectURL == "" {
		c.OIDC.RedirectURL = strings.TrimRight(c.HTTP.PublicBaseURL, "/") + "/api/auth/callback"
	}

	return c, c.validate()
}

func (c *Config) validate() error {
	switch c.DB.Driver {
	case "sqlite", "postgres":
	default:
		return fmt.Errorf("unsupported HP_DB_DRIVER %q (want sqlite or postgres)", c.DB.Driver)
	}
	if c.OIDCEnabled() && c.OIDC.ClientID == "" {
		return fmt.Errorf("HP_OIDC_CLIENT_ID is required when HP_OIDC_ISSUER is set")
	}
	return nil
}

// OIDCEnabled reports whether real SSO is configured; dev-login is used if not.
func (c *Config) OIDCEnabled() bool { return c.OIDC.Issuer != "" }

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

func envDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func splitList(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
