package printing

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/bryansmee/homeprojects/internal/authz"
)

func (e *Extension) authorize(w http.ResponseWriter, r *http.Request, projectID, action string) bool {
	ok, err := e.deps.Authz.AuthorizeProject(r.Context(), projectID, action)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
		return false
	}
	if !ok {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return false
	}
	return true
}

func (e *Extension) list(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	if !e.authorize(w, r, projectID, authz.ActionRead) {
		return
	}
	var links []PrintLink
	if err := e.deps.DB.WithContext(r.Context()).
		Where("project_id = ?", projectID).
		Order("created_at ASC").Find(&links).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, links)
}

type createLinkRequest struct {
	TaskID *string `json:"taskId"`
	Source Source  `json:"source"`
	URL    string  `json:"url"`
	Title  string  `json:"title"`
	Notes  string  `json:"notes"`
}

func (e *Extension) create(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	if !e.authorize(w, r, projectID, authz.ActionWriteExt) {
		return
	}
	var req createLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if req.URL == "" || !req.Source.valid() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "url and a valid source are required"})
		return
	}
	link := PrintLink{
		ID: uuid.NewString(), ProjectID: projectID, TaskID: req.TaskID,
		Source: req.Source, URL: req.URL, Title: req.Title, Notes: req.Notes,
	}
	if err := e.deps.DB.WithContext(r.Context()).Create(&link).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusCreated, link)
}

func (e *Extension) delete(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	if !e.authorize(w, r, projectID, authz.ActionWriteExt) {
		return
	}
	linkID := chi.URLParam(r, "linkID")
	if err := e.deps.DB.WithContext(r.Context()).
		Where("id = ? AND project_id = ?", linkID, projectID).
		Delete(&PrintLink{}).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}
