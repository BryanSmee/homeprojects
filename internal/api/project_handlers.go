package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/bryansmee/homeprojects/internal/auth"
	"github.com/bryansmee/homeprojects/internal/authz"
	"github.com/bryansmee/homeprojects/internal/models"
)

func (s *Server) mountProjectRoutes(api chi.Router) {
	api.Route("/projects", func(pr chi.Router) {
		pr.With(s.requireAuth).Post("/", s.createProject)
		pr.With(s.requireAuth).Get("/", s.listProjects)
		pr.Route("/{projectID}", func(one chi.Router) {
			one.Get("/", s.getProject)
			one.Patch("/", s.updateProject)
			one.Delete("/", s.deleteProject)
			one.Patch("/visibility", s.setVisibility)
			s.mountMemberRoutes(one)
			s.mountTaskRoutes(one)
		})
	})
}

type projectInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var in projectInput
	if !decode(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	p := models.Project{
		Name: in.Name, Description: in.Description,
		OwnerID: auth.FromContext(r.Context()).UserID,
	}
	if err := s.store.CreateProject(r.Context(), &p); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toProjectDTO(p, models.ProjectWaiting))
}

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	userID := auth.FromContext(r.Context()).UserID
	projects, err := s.store.ListProjectsForUser(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	out := make([]projectDTO, 0, len(projects))
	for _, p := range projects {
		status, _ := s.store.ProjectStatus(r.Context(), p.ID)
		out = append(out, toProjectDTO(p, status))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) getProject(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorize(w, r, chi.URLParam(r, "projectID"), authz.ActionRead)
	if !ok {
		return
	}
	tasks, err := s.store.ListTasks(r.Context(), project.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	status, _ := s.store.ProjectStatus(r.Context(), project.ID)
	detail := projectDetail{projectDTO: toProjectDTO(project, status), Tasks: tasks}
	if s.canSeeMembers(r, project) {
		detail.Members, _ = s.store.ListMembers(r.Context(), project.ID)
	}
	writeJSON(w, http.StatusOK, detail)
}

// canSeeMembers hides the member list (and member emails) from anonymous
// callers viewing a public project.
func (s *Server) canSeeMembers(r *http.Request, project models.Project) bool {
	p := auth.FromContext(r.Context())
	if !p.Authenticated {
		return false
	}
	if p.UserID == project.OwnerID {
		return true
	}
	_, ok := s.store.RoleFor(r.Context(), project.ID, p.UserID)
	return ok
}

func (s *Server) updateProject(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorize(w, r, chi.URLParam(r, "projectID"), authz.ActionUpdateProject)
	if !ok {
		return
	}
	var in projectInput
	if !decode(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Name) != "" {
		project.Name = in.Name
	}
	project.Description = in.Description
	if err := s.store.UpdateProject(r.Context(), &project); err != nil {
		writeStoreError(w, err)
		return
	}
	status, _ := s.store.ProjectStatus(r.Context(), project.ID)
	writeJSON(w, http.StatusOK, toProjectDTO(project, status))
}

func (s *Server) deleteProject(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorize(w, r, chi.URLParam(r, "projectID"), authz.ActionDeleteProject)
	if !ok {
		return
	}
	if err := s.store.DeleteProject(r.Context(), project.ID); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type visibilityInput struct {
	Public bool `json:"public"`
}

func (s *Server) setVisibility(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorize(w, r, chi.URLParam(r, "projectID"), authz.ActionSetVisibility)
	if !ok {
		return
	}
	var in visibilityInput
	if !decode(w, r, &in) {
		return
	}
	project.Public = in.Public
	if err := s.store.UpdateProject(r.Context(), &project); err != nil {
		writeStoreError(w, err)
		return
	}
	status, _ := s.store.ProjectStatus(r.Context(), project.ID)
	writeJSON(w, http.StatusOK, toProjectDTO(project, status))
}
