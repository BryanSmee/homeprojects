package store

import (
	"context"

	"github.com/bryansmee/homeprojects/internal/models"
)

func (s *Store) CreateTask(ctx context.Context, t *models.Task) error {
	t.ID = newID()
	if t.Status == "" {
		t.Status = models.TaskWaiting
	}
	return s.db.WithContext(ctx).Create(t).Error
}

func (s *Store) GetTask(ctx context.Context, id string) (models.Task, error) {
	var t models.Task
	err := s.db.WithContext(ctx).First(&t, "id = ?", id).Error
	return t, translate(err)
}

// ListTasks returns top-level tasks of a project with their subtasks preloaded.
func (s *Store) ListTasks(ctx context.Context, projectID string) ([]models.Task, error) {
	var tasks []models.Task
	err := s.db.WithContext(ctx).
		Preload("Subtasks").
		Where("project_id = ? AND parent_id IS NULL", projectID).
		Order("created_at ASC").
		Find(&tasks).Error
	return tasks, err
}

func (s *Store) UpdateTask(ctx context.Context, t *models.Task) error {
	return s.db.WithContext(ctx).
		Model(t).
		Select("title", "description", "status", "parent_id").
		Updates(map[string]any{
			"title":       t.Title,
			"description": t.Description,
			"status":      t.Status,
			"parent_id":   t.ParentID,
		}).Error
}

func (s *Store) DeleteTask(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.Task{}, "id = ?", id).Error
}

// ParentExists reports whether parentID is a task in the given project.
func (s *Store) ParentExists(ctx context.Context, projectID, parentID string) bool {
	var n int64
	s.db.WithContext(ctx).Model(&models.Task{}).
		Where("id = ? AND project_id = ?", parentID, projectID).
		Count(&n)
	return n > 0
}
