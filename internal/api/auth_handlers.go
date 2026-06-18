package api

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bryansmee/homeprojects/internal/auth"
)

func (s *Server) mountAuthRoutes(api chi.Router) {
	api.Route("/auth", func(a chi.Router) {
		a.Get("/config", s.handleAuthConfig)
		a.Get("/me", s.handleMe)
		a.Post("/logout", s.handleLogout)
		if s.auth != nil {
			a.Get("/login", s.handleLogin)
			a.Get("/callback", s.handleCallback)
		} else {
			a.Post("/dev-login", s.handleDevLogin)
		}
	})
}

// handleAuthConfig advertises which login method the frontend should offer.
func (s *Server) handleAuthConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"oidcEnabled": s.auth != nil})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	p := auth.FromContext(r.Context())
	if !p.Authenticated {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	user, err := s.store.GetUser(r.Context(), p.UserID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) handleLogout(w http.ResponseWriter, _ *http.Request) {
	s.sessions.Clear(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

const stateCookie = "hp_oauth_state"

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := randomState()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not start login")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: stateCookie, Value: state, Path: "/", MaxAge: 600,
		HttpOnly: true, Secure: s.cfg.Session.CookieSecure, SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, s.auth.AuthCodeURL(state), http.StatusFound)
}

func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(stateCookie)
	if err != nil || cookie.Value == "" || cookie.Value != r.URL.Query().Get("state") {
		writeError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}
	claims, err := s.auth.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "login failed")
		return
	}
	user, err := s.store.UpsertUser(r.Context(), claims.Subject, claims.Email, claims.Name)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	s.issueSession(w, user.ID, claims.Subject, claims.Email, claims.Name)
	http.Redirect(w, r, s.cfg.HTTP.FrontendBaseURL, http.StatusFound)
}

type devLoginRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// handleDevLogin issues a session without an IdP. Only mounted when OIDC is off.
func (s *Server) handleDevLogin(w http.ResponseWriter, r *http.Request) {
	var req devLoginRequest
	if !decode(w, r, &req) {
		return
	}
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}
	subject := "dev|" + req.Email
	user, err := s.store.UpsertUser(r.Context(), subject, req.Email, req.Name)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	s.issueSession(w, user.ID, subject, req.Email, req.Name)
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) issueSession(w http.ResponseWriter, id, subject, email, name string) {
	_ = s.sessions.Issue(w, auth.Principal{
		UserID: id, Subject: subject, Email: email, Name: name, Authenticated: true,
	})
}

func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
