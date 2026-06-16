package store

import (
	"context"

	"gorm.io/gorm/clause"

	"github.com/bryansmee/homeprojects/internal/models"
)

func (s *Store) ListMembers(ctx context.Context, projectID string) ([]models.Membership, error) {
	var members []models.Membership
	err := s.db.WithContext(ctx).
		Preload("User").
		Where("project_id = ?", projectID).
		Order("created_at ASC").
		Find(&members).Error
	return members, err
}

// UpsertMember adds a member or updates their role if already present.
func (s *Store) UpsertMember(ctx context.Context, projectID, userID string, role models.Role) (models.Membership, error) {
	m := models.Membership{ID: newID(), ProjectID: projectID, UserID: userID, Role: role}
	err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role", "updated_at"}),
	}).Create(&m).Error
	return m, err
}

func (s *Store) RemoveMember(ctx context.Context, projectID, userID string) error {
	return s.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Delete(&models.Membership{}).Error
}

func (s *Store) GetMember(ctx context.Context, projectID, userID string) (models.Membership, error) {
	var m models.Membership
	err := s.db.WithContext(ctx).Preload("User").
		First(&m, "project_id = ? AND user_id = ?", projectID, userID).Error
	return m, translate(err)
}

// CountAdmins is used to prevent removing the last admin of a project.
func (s *Store) CountAdmins(ctx context.Context, projectID string) (int64, error) {
	var n int64
	err := s.db.WithContext(ctx).Model(&models.Membership{}).
		Where("project_id = ? AND role = ?", projectID, models.RoleAdmin).
		Count(&n).Error
	return n, err
}
