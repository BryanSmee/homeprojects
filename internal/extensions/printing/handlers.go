package printing

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/bryansmee/homeprojects/internal/auth"
	"github.com/bryansmee/homeprojects/internal/authz"
	"github.com/bryansmee/homeprojects/internal/models"
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

// list returns all of a project's links (the files overview), optionally
// filtered to a single task via ?taskId=.
func (e *Extension) list(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	if !e.authorize(w, r, projectID, authz.ActionRead) {
		return
	}
	q := e.deps.DB.WithContext(r.Context()).Where("project_id = ?", projectID)
	if taskID := r.URL.Query().Get("taskId"); taskID != "" {
		q = q.Where("task_id = ?", taskID)
	}
	var links []PrintLink
	if err := q.Order("created_at ASC").Find(&links).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, links)
}

// preview resolves a URL's OpenGraph title/thumbnail so the UI can prefill the
// add form. Requires authentication (it makes outbound requests on our behalf).
func (e *Extension) preview(w http.ResponseWriter, r *http.Request) {
	if !auth.FromContext(r.Context()).Authenticated {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	target := r.URL.Query().Get("url")
	if target == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "url is required"})
		return
	}
	preview, err := resolveLinkPreview(r.Context(), target)
	if err != nil {
		// Not fatal for the client; just means we couldn't find a preview.
		writeJSON(w, http.StatusOK, linkPreview{})
		return
	}
	writeJSON(w, http.StatusOK, preview)
}

// taskInProject reports whether taskID belongs to projectID.
func (e *Extension) taskInProject(r *http.Request, projectID, taskID string) bool {
	var n int64
	e.deps.DB.WithContext(r.Context()).Model(&models.Task{}).
		Where("id = ? AND project_id = ?", taskID, projectID).
		Count(&n)
	return n > 0
}

type createLinkRequest struct {
	TaskID       string `json:"taskId"`
	Source       Source `json:"source"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Title        string `json:"title"`
	Notes        string `json:"notes"`
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
	if req.TaskID == "" || !e.taskInProject(r, projectID, req.TaskID) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "a valid taskId in this project is required"})
		return
	}
	// Best-effort: derive a thumbnail (and title) from the page when not given.
	if req.ThumbnailURL == "" || req.Title == "" {
		if preview, err := resolveLinkPreview(r.Context(), req.URL); err == nil {
			if req.ThumbnailURL == "" {
				req.ThumbnailURL = preview.ThumbnailURL
			}
			if req.Title == "" {
				req.Title = preview.Title
			}
		}
	}
	link := PrintLink{
		ID: uuid.NewString(), ProjectID: projectID, TaskID: req.TaskID,
		Source: req.Source, URL: req.URL, ThumbnailURL: req.ThumbnailURL,
		Title: req.Title, Notes: req.Notes,
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
