// Package extensions defines a plugin system: self-contained features that
// register their own models and HTTP routes, reusing the core authorization.
package extensions

import (
	"context"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// Authorizer lets extensions reuse the core project-level access checks. The
// caller's principal is taken from the request context.
type Authorizer interface {
	AuthorizeProject(ctx context.Context, projectID, action string) (bool, error)
}

// Deps are the host-provided dependencies handed to each extension.
type Deps struct {
	DB    *gorm.DB
	Authz Authorizer
}

// Extension is a pluggable feature module.
type Extension interface {
	Name() string
	Models() []any
	Mount(r chi.Router, deps Deps)
}

// Registry collects the extensions enabled in a build.
type Registry struct {
	exts []Extension
}

func NewRegistry(exts ...Extension) *Registry { return &Registry{exts: exts} }

func (reg *Registry) All() []Extension { return reg.exts }

func (reg *Registry) Models() []any {
	var m []any
	for _, e := range reg.exts {
		m = append(m, e.Models()...)
	}
	return m
}
