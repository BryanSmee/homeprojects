package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bryansmee/homeprojects/internal/authz"
	"github.com/bryansmee/homeprojects/internal/models"
)

func (s *Server) mountMemberRoutes(one chi.Router) {
	one.Route("/members", func(m chi.Router) {
		m.Get("/", s.listMembers)
		m.Post("/", s.addMember)
		m.Delete("/{userID}", s.removeMember)
	})
}

func (s *Server) listMembers(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorize(w, r, chi.URLParam(r, "projectID"), authz.ActionRead)
	if !ok {
		return
	}
	if !s.canSeeMembers(r, project) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	members, err := s.store.ListMembers(r.Context(), project.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, members)
}

type memberInput struct {
	UserID string      `json:"userId"`
	Email  string      `json:"email"`
	Role   models.Role `json:"role"`
}

func (s *Server) addMember(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorize(w, r, chi.URLParam(r, "projectID"), authz.ActionManageMembers)
	if !ok {
		return
	}
	var in memberInput
	if !decode(w, r, &in) {
		return
	}
	if !in.Role.Valid() {
		writeError(w, http.StatusBadRequest, "invalid role")
		return
	}
	user, err := s.resolveUser(r, in)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found (they must sign in once first)")
		return
	}
	member, err := s.store.UpsertMember(r.Context(), project.ID, user.ID, in.Role)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	member.User = &user
	writeJSON(w, http.StatusOK, member)
}

func (s *Server) resolveUser(r *http.Request, in memberInput) (models.User, error) {
	if in.UserID != "" {
		return s.store.GetUser(r.Context(), in.UserID)
	}
	return s.store.GetUserByEmail(r.Context(), in.Email)
}

func (s *Server) removeMember(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorize(w, r, chi.URLParam(r, "projectID"), authz.ActionManageMembers)
	if !ok {
		return
	}
	userID := chi.URLParam(r, "userID")
	if s.wouldOrphanProject(r, project, userID) {
		writeError(w, http.StatusConflict, "cannot remove the last admin")
		return
	}
	if err := s.store.RemoveMember(r.Context(), project.ID, userID); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// wouldOrphanProject reports whether removing userID leaves the project adminless.
func (s *Server) wouldOrphanProject(r *http.Request, project models.Project, userID string) bool {
	member, err := s.store.GetMember(r.Context(), project.ID, userID)
	if err != nil || member.Role != models.RoleAdmin {
		return false
	}
	admins, err := s.store.CountAdmins(r.Context(), project.ID)
	return err == nil && admins <= 1
}
