package authz

import (
	"context"
	"testing"
)

func TestPolicy(t *testing.T) {
	eng, err := New(context.Background())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	owner := Subject{ID: "u1", Authenticated: true}
	editor := Subject{ID: "u2", Authenticated: true}
	stranger := Subject{ID: "u3", Authenticated: true}
	anon := Subject{Authenticated: false}
	priv := ProjectContext{Public: false, OwnerID: "u1"}
	pub := ProjectContext{Public: true, OwnerID: "u1"}

	tests := []struct {
		name string
		in   Input
		want bool
	}{
		{"owner deletes project", Input{ActionDeleteProject, owner, priv}, true},
		{"editor updates task", Input{ActionUpdateTask, editor, withRole(priv, "editor")}, true},
		{"editor cannot manage members", Input{ActionManageMembers, editor, withRole(priv, "editor")}, false},
		{"viewer reads", Input{ActionRead, editor, withRole(priv, "viewer")}, true},
		{"viewer cannot write", Input{ActionUpdateTask, editor, withRole(priv, "viewer")}, false},
		{"stranger cannot read private", Input{ActionRead, stranger, priv}, false},
		{"anon reads public", Input{ActionRead, anon, pub}, true},
		{"anon cannot write public", Input{ActionWriteExt, anon, pub}, false},
		{"any auth user creates project", Input{ActionCreateProject, stranger, ProjectContext{}}, true},
		{"anon cannot create project", Input{ActionCreateProject, anon, ProjectContext{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := eng.Allow(context.Background(), tt.in)
			if err != nil {
				t.Fatalf("Allow: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func withRole(p ProjectContext, role string) ProjectContext {
	p.Role = role
	return p
}
