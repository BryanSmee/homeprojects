package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SessionManager issues and verifies signed session cookies. The cookie holds a
// short JWT identifying the user; no server-side session store is required,
// which keeps the backend stateless and easy to scale horizontally on K8s.
type SessionManager struct {
	secret     []byte
	ttl        time.Duration
	cookieName string
	secure     bool
	domain     string
}

// NewSessionManager constructs a SessionManager.
func NewSessionManager(secret string, ttl time.Duration, cookieName string, secure bool, domain string) *SessionManager {
	return &SessionManager{
		secret:     []byte(secret),
		ttl:        ttl,
		cookieName: cookieName,
		secure:     secure,
		domain:     domain,
	}
}

type sessionClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Issue writes a session cookie for the given user to w.
func (m *SessionManager) Issue(w http.ResponseWriter, p Principal) error {
	now := time.Now()
	claims := sessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   p.UserID,
			ID:        p.Subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
		Email: p.Email,
		Name:  p.Name,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return fmt.Errorf("sign session: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    signed,
		Path:     "/",
		Domain:   m.domain,
		Expires:  now.Add(m.ttl),
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// Clear removes the session cookie.
func (m *SessionManager) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
		Domain:   m.domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// Parse reads and verifies the session cookie from r, returning the principal.
func (m *SessionManager) Parse(r *http.Request) (Principal, bool) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return Anonymous, false
	}

	var claims sessionClaims
	_, err = jwt.ParseWithClaims(cookie.Value, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return Anonymous, false
	}

	return Principal{
		UserID:        claims.Subject,
		Subject:       claims.ID,
		Email:         claims.Email,
		Name:          claims.Name,
		Authenticated: true,
	}, true
}
