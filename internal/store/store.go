// Package store provides persistence operations over GORM for the core models.
package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/bryansmee/homeprojects/internal/models"
)

var ErrNotFound = errors.New("not found")

type Store struct{ db *gorm.DB }

func New(db *gorm.DB) *Store { return &Store{db: db} }

// DB exposes the underlying connection for extensions that manage their own models.
func (s *Store) DB() *gorm.DB { return s.db }

func newID() string { return uuid.NewString() }

func translate(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *Store) UpsertUser(ctx context.Context, subject, email, name string) (models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).Where("subject = ?", subject).First(&user).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		user = models.User{ID: newID(), Subject: subject, Email: email, Name: name}
		return user, s.db.WithContext(ctx).Create(&user).Error
	case err != nil:
		return user, err
	}

	user.Email, user.Name = email, name
	return user, s.db.WithContext(ctx).Save(&user).Error
}

func (s *Store) GetUser(ctx context.Context, id string) (models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).First(&user, "id = ?", id).Error
	return user, translate(err)
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).First(&user, "email = ?", email).Error
	return user, translate(err)
}
