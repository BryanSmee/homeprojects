package models

import "testing"

func TestDeriveProjectStatus(t *testing.T) {
	tests := []struct {
		name string
		in   []TaskStatus
		want ProjectStatus
	}{
		{"no tasks", nil, ProjectWaiting},
		{"any in progress wins", []TaskStatus{TaskDone, TaskInProgress, TaskWaiting}, ProjectInProgress},
		{"all abandoned", []TaskStatus{TaskAbandoned, TaskAbandoned}, ProjectAbandoned},
		{"done and abandoned", []TaskStatus{TaskDone, TaskAbandoned}, ProjectDone},
		{"all done", []TaskStatus{TaskDone, TaskDone}, ProjectDone},
		{"some waiting", []TaskStatus{TaskDone, TaskWaiting}, ProjectWaiting},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeriveProjectStatus(tt.in); got != tt.want {
				t.Errorf("DeriveProjectStatus(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRoleAtLeast(t *testing.T) {
	if !RoleAdmin.AtLeast(RoleEditor) {
		t.Error("admin should be at least editor")
	}
	if RoleViewer.AtLeast(RoleEditor) {
		t.Error("viewer should not be at least editor")
	}
}
