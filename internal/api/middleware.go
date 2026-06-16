package api

import (
	"net/http"

	"github.com/bryansmee/homeprojects/internal/auth"
)

// sessionMiddleware parses the session cookie (if any) into the request context.
// It never rejects; enforcement is done by requireAuth and policy checks.
func (s *Server) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p, ok := s.sessions.Parse(r); ok {
			r = r.WithContext(auth.WithPrincipal(r.Context(), p))
		}
		next.ServeHTTP(w, r)
	})
}

// requireAuth rejects unauthenticated requests with 401.
func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.FromContext(r.Context()).Authenticated {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
