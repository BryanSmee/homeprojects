package api

import (
	"context"
	"net/http"

	"github.com/bryansmee/homeprojects/internal/auth"
	"github.com/bryansmee/homeprojects/internal/authz"
	"github.com/bryansmee/homeprojects/internal/models"
)

// buildInput assembles a policy input for a known project and principal.
func (s *Server) buildInput(ctx context.Context, p auth.Principal, project models.Project, action string) authz.Input {
	role := ""
	if p.Authenticated {
		if r, ok := s.store.RoleFor(ctx, project.ID, p.UserID); ok {
			role = string(r)
		}
	}
	return authz.Input{
		Action:  action,
		Subject: authz.Subject{ID: p.UserID, Authenticated: p.Authenticated},
		Project: authz.ProjectContext{Public: project.Public, OwnerID: project.OwnerID, Role: role},
	}
}

// AuthorizeProject loads the project and evaluates the policy. It satisfies
// extensions.Authorizer so extensions reuse the same access rules.
func (s *Server) AuthorizeProject(ctx context.Context, projectID, action string) (bool, error) {
	p := auth.FromContext(ctx)
	project, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return false, err
	}
	return s.authz.Allow(ctx, s.buildInput(ctx, p, project, action))
}

// authorize is the in-handler guard: it loads the project, checks the policy,
// and on success returns the project for reuse. It writes the error response
// itself and returns ok=false when the caller must stop.
func (s *Server) authorize(w http.ResponseWriter, r *http.Request, projectID, action string) (models.Project, bool) {
	project, err := s.store.GetProject(r.Context(), projectID)
	if err != nil {
		writeStoreError(w, err)
		return project, false
	}
	p := auth.FromContext(r.Context())
	allowed, err := s.authz.Allow(r.Context(), s.buildInput(r.Context(), p, project, action))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "authorization error")
		return project, false
	}
	if !allowed {
		s.denied(w, p)
		return project, false
	}
	return project, true
}

// denied returns 401 for anonymous callers and 403 for authenticated ones.
func (s *Server) denied(w http.ResponseWriter, p auth.Principal) {
	if !p.Authenticated {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	writeError(w, http.StatusForbidden, "forbidden")
}
