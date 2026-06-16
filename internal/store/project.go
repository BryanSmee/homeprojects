package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/bryansmee/homeprojects/internal/models"
)

// CreateProject persists p and grants the owner an admin membership.
func (s *Store) CreateProject(ctx context.Context, p *models.Project) error {
	p.ID = newID()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		owner := models.Membership{
			ID: newID(), ProjectID: p.ID, UserID: p.OwnerID, Role: models.RoleAdmin,
		}
		return tx.Create(&owner).Error
	})
}

func (s *Store) GetProject(ctx context.Context, id string) (models.Project, error) {
	var p models.Project
	err := s.db.WithContext(ctx).First(&p, "id = ?", id).Error
	return p, translate(err)
}

// ListProjectsForUser returns projects the user owns or is a member of.
func (s *Store) ListProjectsForUser(ctx context.Context, userID string) ([]models.Project, error) {
	var projects []models.Project
	err := s.db.WithContext(ctx).
		Joins("LEFT JOIN memberships m ON m.project_id = projects.id AND m.user_id = ?", userID).
		Where("projects.owner_id = ? OR m.user_id = ?", userID, userID).
		Group("projects.id").
		Order("projects.created_at DESC").
		Find(&projects).Error
	return projects, err
}

func (s *Store) UpdateProject(ctx context.Context, p *models.Project) error {
	return s.db.WithContext(ctx).
		Model(p).
		Select("name", "description", "public").
		Updates(p).Error
}

func (s *Store) DeleteProject(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.Project{}, "id = ?", id).Error
}

// RoleFor returns the user's role on a project, if any.
func (s *Store) RoleFor(ctx context.Context, projectID, userID string) (models.Role, bool) {
	var m models.Membership
	err := s.db.WithContext(ctx).
		First(&m, "project_id = ? AND user_id = ?", projectID, userID).Error
	if err != nil {
		return "", false
	}
	return m.Role, true
}

// ProjectStatus derives a project's status from its task statuses.
func (s *Store) ProjectStatus(ctx context.Context, projectID string) (models.ProjectStatus, error) {
	var statuses []models.TaskStatus
	err := s.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("project_id = ?", projectID).
		Pluck("status", &statuses).Error
	if err != nil {
		return "", err
	}
	return models.DeriveProjectStatus(statuses), nil
}
