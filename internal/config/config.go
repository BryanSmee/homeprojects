// Package config loads runtime configuration from environment variables, with
// dev-friendly defaults so the server boots with zero config against SQLite.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds all runtime configuration for the server.
type Config struct {
	// HTTP
	Addr           string
	AllowedOrigins []string
	PublicBaseURL  string // public URL of this API, used to build OIDC redirect URLs

	// Driver is a registered DB driver name (see internal/db).
	DBDriver string
	DBDSN    string

	// OIDC/SSO (e.g. Pocket ID). Empty issuer enables dev-login mode.
	OIDCIssuer       string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string
	OIDCScopes       []string

	// Session signing key for the JWT stored in the session cookie.
	SessionSecret   string
	SessionTTL      time.Duration
	CookieName      string
	CookieSecure    bool
	CookieDomain    string
	FrontendBaseURL string
}

// Load reads configuration from the environment, applying defaults.
func Load() (*Config, error) {
	c := &Config{
		Addr:             env("HP_ADDR", ":8080"),
		AllowedOrigins:   splitList(env("HP_ALLOWED_ORIGINS", "http://localhost:3000")),
		PublicBaseURL:    env("HP_PUBLIC_BASE_URL", "http://localhost:8080"),
		DBDriver:         env("HP_DB_DRIVER", "sqlite"),
		DBDSN:            env("HP_DB_DSN", "homeprojects.db"),
		OIDCIssuer:       env("HP_OIDC_ISSUER", ""),
		OIDCClientID:     env("HP_OIDC_CLIENT_ID", ""),
		OIDCClientSecret: env("HP_OIDC_CLIENT_SECRET", ""),
		OIDCRedirectURL:  env("HP_OIDC_REDIRECT_URL", ""),
		OIDCScopes:       splitList(env("HP_OIDC_SCOPES", "openid,profile,email")),
		SessionSecret:    env("HP_SESSION_SECRET", "dev-insecure-change-me"),
		SessionTTL:       envDuration("HP_SESSION_TTL", 7*24*time.Hour),
		CookieName:       env("HP_COOKIE_NAME", "hp_session"),
		CookieSecure:     envBool("HP_COOKIE_SECURE", false),
		CookieDomain:     env("HP_COOKIE_DOMAIN", ""),
		FrontendBaseURL:  env("HP_FRONTEND_BASE_URL", "http://localhost:3000"),
	}

	if c.OIDCRedirectURL == "" {
		c.OIDCRedirectURL = strings.TrimRight(c.PublicBaseURL, "/") + "/api/auth/callback"
	}

	return c, c.validate()
}

func (c *Config) validate() error {
	switch c.DBDriver {
	case "sqlite", "postgres":
	default:
		return fmt.Errorf("unsupported HP_DB_DRIVER %q (want sqlite or postgres)", c.DBDriver)
	}
	if c.OIDCEnabled() {
		if c.OIDCClientID == "" {
			return fmt.Errorf("HP_OIDC_CLIENT_ID is required when HP_OIDC_ISSUER is set")
		}
	}
	return nil
}

func (c *Config) OIDCEnabled() bool { return c.OIDCIssuer != "" }

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
