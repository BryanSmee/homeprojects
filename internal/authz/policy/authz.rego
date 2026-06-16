package homeprojects.authz

import rego.v1

default allow := false

# Anyone, including anonymous callers, may read a public project.
allow if {
	input.action == "read"
	input.project.public == true
}

# Any authenticated user may create a project.
allow if {
	input.action == "create_project"
	input.subject.authenticated
}

# A project owner has full control over their project.
allow if {
	input.subject.authenticated
	input.subject.id == input.project.owner_id
}

# Otherwise, permission is granted by the caller's role on the project.
allow if {
	input.subject.authenticated
	role_actions[input.project.role][input.action]
}

role_actions := {
	"viewer": {"read"},
	"editor": {"read", "create_task", "update_task", "delete_task", "write_extension"},
	"admin": {
		"read", "create_task", "update_task", "delete_task", "write_extension",
		"update_project", "delete_project", "manage_members", "set_visibility",
	},
}
