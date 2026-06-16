package api

import (
	"time"

	"github.com/bryansmee/homeprojects/internal/models"
)

type projectDTO struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	OwnerID     string               `json:"ownerId"`
	Public      bool                 `json:"public"`
	Status      models.ProjectStatus `json:"status"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
}

func toProjectDTO(p models.Project, status models.ProjectStatus) projectDTO {
	return projectDTO{
		ID: p.ID, Name: p.Name, Description: p.Description, OwnerID: p.OwnerID,
		Public: p.Public, Status: status, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

type projectDetail struct {
	projectDTO
	Tasks   []models.Task       `json:"tasks"`
	Members []models.Membership `json:"members,omitempty"`
}
