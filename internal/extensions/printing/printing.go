// Package printing is a 3D-printing extension: it attaches links to external
// model sources (Thingiverse, Printables, Cults3D, ...) to a project's tasks.
// Printer/print-job status is intended to be added here later.
package printing

import (
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/bryansmee/homeprojects/internal/extensions"
)

type Source string

const (
	SourceThingiverse Source = "thingiverse"
	SourcePrintables  Source = "printables"
	SourceCults3D     Source = "cults3d"
	SourceMakerWorld  Source = "makerworld"
	SourceOther       Source = "other"
)

func (s Source) valid() bool {
	switch s {
	case SourceThingiverse, SourcePrintables, SourceCults3D, SourceMakerWorld, SourceOther:
		return true
	default:
		return false
	}
}

// PrintLink references an external 3D model, tied to a task within a project.
// ProjectID is denormalised from the task to scope authorization and to power
// the project-wide files overview.
type PrintLink struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	ProjectID    string    `gorm:"index;size:36;not null" json:"projectId"`
	TaskID       string    `gorm:"index;size:36;not null" json:"taskId"`
	Source       Source    `gorm:"size:32;not null" json:"source"`
	URL          string    `gorm:"size:1024;not null" json:"url"`
	ThumbnailURL string    `gorm:"size:1024" json:"thumbnailUrl"`
	Title        string    `gorm:"size:255" json:"title"`
	Notes        string    `gorm:"type:text" json:"notes"`
	Status       string    `gorm:"size:32" json:"status,omitempty"` // reserved for printer status
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Extension struct{ deps extensions.Deps }

func New() *Extension { return &Extension{} }

func (e *Extension) Name() string  { return "printing" }
func (e *Extension) Models() []any { return []any{&PrintLink{}} }

func (e *Extension) Mount(r chi.Router, deps extensions.Deps) {
	e.deps = deps
	r.Get("/projects/{projectID}/links", e.list)
	r.Post("/projects/{projectID}/links", e.create)
	r.Delete("/projects/{projectID}/links/{linkID}", e.delete)
}
