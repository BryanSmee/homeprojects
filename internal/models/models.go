// Package models defines the persistent domain entities and the business rules
// that operate purely on them (e.g. project status derivation).
package models

import (
	"time"
)

// Role is a member's role within a project.
type Role string

const (
	RoleAdmin  Role = "admin"  // full control incl. members, visibility, deletion
	RoleEditor Role = "editor" // create/update/delete tasks
	RoleViewer Role = "viewer" // read-only
)

// Valid reports whether r is a recognised role.
func (r Role) Valid() bool {
	switch r {
	case RoleAdmin, RoleEditor, RoleViewer:
		return true
	default:
		return false
	}
}

// rank gives roles a comparable strength for "at least" checks.
func (r Role) rank() int {
	switch r {
	case RoleViewer:
		return 1
	case RoleEditor:
		return 2
	case RoleAdmin:
		return 3
	default:
		return 0
	}
}

// AtLeast reports whether r is at least as powerful as other.
func (r Role) AtLeast(other Role) bool { return r.rank() >= other.rank() }

// User is a person authenticated via SSO. Subject is the stable OIDC `sub`
// claim and is the primary identity key.
type User struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	Subject   string    `gorm:"uniqueIndex;size:255" json:"-"`
	Email     string    `gorm:"index;size:320" json:"email"`
	Name      string    `gorm:"size:255" json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Project groups tasks. Visibility is private by default; setting Public to
// true allows anyone with the link to read it without authentication.
type Project struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	OwnerID     string    `gorm:"index;size:36;not null" json:"ownerId"`
	Public      bool      `gorm:"not null;default:false" json:"public"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Members []Membership `gorm:"constraint:OnDelete:CASCADE;" json:"members,omitempty"`
	Tasks   []Task       `gorm:"constraint:OnDelete:CASCADE;" json:"tasks,omitempty"`
}

// Membership grants a user a role on a project.
type Membership struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	ProjectID string    `gorm:"index:idx_member_project_user,unique;size:36;not null" json:"projectId"`
	UserID    string    `gorm:"index:idx_member_project_user,unique;size:36;not null" json:"userId"`
	Role      Role      `gorm:"size:16;not null" json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	User *User `gorm:"constraint:OnDelete:CASCADE;" json:"user,omitempty"`
}

// Task is a unit of work in a project. A subtask is simply a task whose
// ParentID points at another task in the same project.
type Task struct {
	ID          string     `gorm:"primaryKey;size:36" json:"id"`
	ProjectID   string     `gorm:"index;size:36;not null" json:"projectId"`
	ParentID    *string    `gorm:"index;size:36" json:"parentId,omitempty"`
	Title       string     `gorm:"size:255;not null" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	Status      TaskStatus `gorm:"size:16;not null;default:'Waiting'" json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`

	Subtasks []Task `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE;" json:"subtasks,omitempty"`
}

// AllModels returns every core model for auto-migration. Extensions register
// their own models separately via the extension registry.
func AllModels() []any {
	return []any{
		&User{},
		&Project{},
		&Membership{},
		&Task{},
	}
}
