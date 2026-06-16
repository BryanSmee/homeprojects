package models

// TaskStatus enumerates the lifecycle states a task (or subtask) can be in.
type TaskStatus string

const (
	TaskWaiting    TaskStatus = "Waiting"
	TaskInProgress TaskStatus = "In Progress"
	TaskDone       TaskStatus = "Done"
	TaskAbandoned  TaskStatus = "Abandoned"
)

// Valid reports whether s is a recognised task status.
func (s TaskStatus) Valid() bool {
	switch s {
	case TaskWaiting, TaskInProgress, TaskDone, TaskAbandoned:
		return true
	default:
		return false
	}
}

// ProjectStatus is derived from a project's tasks; it is never stored.
type ProjectStatus string

const (
	ProjectWaiting    ProjectStatus = "Waiting"
	ProjectInProgress ProjectStatus = "In Progress"
	ProjectDone       ProjectStatus = "Done"
	ProjectAbandoned  ProjectStatus = "Abandoned"
)

// DeriveProjectStatus computes a project's status from the statuses of all of
// its tasks (top-level tasks and subtasks alike), following these rules:
//
//   - no tasks                       -> Waiting (nothing to do yet)
//   - any task In Progress           -> In Progress
//   - every task Abandoned           -> Abandoned
//   - every task Done or Abandoned   -> Done
//   - otherwise (some still Waiting) -> Waiting
func DeriveProjectStatus(statuses []TaskStatus) ProjectStatus {
	if len(statuses) == 0 {
		return ProjectWaiting
	}

	var done, abandoned int
	for _, s := range statuses {
		switch s {
		case TaskInProgress:
			return ProjectInProgress
		case TaskDone:
			done++
		case TaskAbandoned:
			abandoned++
		}
	}

	switch {
	case abandoned == len(statuses):
		return ProjectAbandoned
	case done+abandoned == len(statuses):
		return ProjectDone
	default:
		return ProjectWaiting
	}
}
