// Package authz evaluates access decisions using an embedded OPA/Rego policy.
package authz

import (
	"context"
	"embed"
	"fmt"

	"github.com/open-policy-agent/opa/rego"
)

// Action names referenced by the Rego policy.
const (
	ActionRead          = "read"
	ActionCreateProject = "create_project"
	ActionUpdateProject = "update_project"
	ActionDeleteProject = "delete_project"
	ActionSetVisibility = "set_visibility"
	ActionManageMembers = "manage_members"
	ActionCreateTask    = "create_task"
	ActionUpdateTask    = "update_task"
	ActionDeleteTask    = "delete_task"
	ActionWriteExt      = "write_extension"
)

type Subject struct {
	ID            string `json:"id"`
	Authenticated bool   `json:"authenticated"`
}

type ProjectContext struct {
	Public  bool   `json:"public"`
	OwnerID string `json:"owner_id"`
	Role    string `json:"role"` // caller's role on the project, "" if none
}

type Input struct {
	Action  string         `json:"action"`
	Subject Subject        `json:"subject"`
	Project ProjectContext `json:"project"`
}

//go:embed policy/authz.rego
var policyFS embed.FS

// Engine holds a prepared Rego query for fast repeated evaluation.
type Engine struct {
	query rego.PreparedEvalQuery
}

func New(ctx context.Context) (*Engine, error) {
	module, err := policyFS.ReadFile("policy/authz.rego")
	if err != nil {
		return nil, fmt.Errorf("read embedded policy: %w", err)
	}
	q, err := rego.New(
		rego.Query("data.homeprojects.authz.allow"),
		rego.Module("authz.rego", string(module)),
	).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("prepare policy: %w", err)
	}
	return &Engine{query: q}, nil
}

func (e *Engine) Allow(ctx context.Context, in Input) (bool, error) {
	rs, err := e.query.Eval(ctx, rego.EvalInput(in))
	if err != nil {
		return false, fmt.Errorf("eval policy: %w", err)
	}
	return rs.Allowed(), nil
}
