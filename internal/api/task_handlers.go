package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/bryansmee/homeprojects/internal/authz"
	"github.com/bryansmee/homeprojects/internal/models"
)

func (s *Server) mountTaskRoutes(one chi.Router) {
	one.Route("/tasks", func(t chi.Router) {
		t.Get("/", s.listTasks)
		t.Post("/", s.createTask)
		t.Patch("/{taskID}", s.updateTask)
		t.Delete("/{taskID}", s.deleteTask)
	})
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorize(w, r, chi.URLParam(r, "projectID"), authz.ActionRead)
	if !ok {
		return
	}
	tasks, err := s.store.ListTasks(r.Context(), project.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

type taskInput struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      models.TaskStatus `json:"status"`
	ParentID    *string           `json:"parentId"`
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorize(w, r, chi.URLParam(r, "projectID"), authz.ActionCreateTask)
	if !ok {
		return
	}
	var in taskInput
	if !decode(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if in.Status != "" && !in.Status.Valid() {
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}
	if !s.validParent(r, project.ID, in.ParentID) {
		writeError(w, http.StatusBadRequest, "parent task not found in this project")
		return
	}
	task := models.Task{
		ProjectID: project.ID, Title: in.Title, Description: in.Description,
		Status: in.Status, ParentID: in.ParentID,
	}
	if err := s.store.CreateTask(r.Context(), &task); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (s *Server) updateTask(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorize(w, r, chi.URLParam(r, "projectID"), authz.ActionUpdateTask)
	if !ok {
		return
	}
	task, ok := s.taskInProject(w, r, project.ID)
	if !ok {
		return
	}
	var in taskInput
	if !decode(w, r, &in) {
		return
	}
	if in.Status != "" && !in.Status.Valid() {
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}
	applyTaskUpdate(&task, in)
	if err := s.store.UpdateTask(r.Context(), &task); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func applyTaskUpdate(task *models.Task, in taskInput) {
	if strings.TrimSpace(in.Title) != "" {
		task.Title = in.Title
	}
	task.Description = in.Description
	if in.Status != "" {
		task.Status = in.Status
	}
}

func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request) {
	project, ok := s.authorize(w, r, chi.URLParam(r, "projectID"), authz.ActionDeleteTask)
	if !ok {
		return
	}
	task, ok := s.taskInProject(w, r, project.ID)
	if !ok {
		return
	}
	if err := s.store.DeleteTask(r.Context(), task.ID); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// taskInProject loads the task named in the URL and verifies it belongs to the
// project, preventing cross-project access via a guessed task ID.
func (s *Server) taskInProject(w http.ResponseWriter, r *http.Request, projectID string) (models.Task, bool) {
	task, err := s.store.GetTask(r.Context(), chi.URLParam(r, "taskID"))
	if err != nil {
		writeStoreError(w, err)
		return task, false
	}
	if task.ProjectID != projectID {
		writeError(w, http.StatusNotFound, "not found")
		return task, false
	}
	return task, true
}

func (s *Server) validParent(r *http.Request, projectID string, parentID *string) bool {
	if parentID == nil || *parentID == "" {
		return true
	}
	return s.store.ParentExists(r.Context(), projectID, *parentID)
}
