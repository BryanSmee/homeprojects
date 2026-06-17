"use client";

import * as React from "react";
import { toast } from "sonner";
import { Plus, Trash2, CornerDownRight } from "lucide-react";

import { api } from "@/lib/api";
import { TASK_STATUSES, type Task, type TaskStatus } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface Props {
  projectId: string;
  tasks: Task[];
  canEdit: boolean;
  onChange: () => void;
}

export function TaskList({ projectId, tasks, canEdit, onChange }: Props) {
  return (
    <div className="flex flex-col gap-2">
      {canEdit && <AddTaskForm projectId={projectId} onChange={onChange} />}
      {tasks.length === 0 ? (
        <p className="py-6 text-center text-sm text-muted-foreground">
          No tasks yet.
        </p>
      ) : (
        tasks.map((task) => (
          <TaskRow
            key={task.id}
            projectId={projectId}
            task={task}
            canEdit={canEdit}
            onChange={onChange}
          />
        ))
      )}
    </div>
  );
}

function TaskRow({
  projectId,
  task,
  canEdit,
  onChange,
}: {
  projectId: string;
  task: Task;
  canEdit: boolean;
  onChange: () => void;
}) {
  const [addingSub, setAddingSub] = React.useState(false);

  return (
    <div className="rounded-lg border">
      <TaskRowInner
        projectId={projectId}
        task={task}
        canEdit={canEdit}
        onChange={onChange}
        onAddSub={() => setAddingSub((v) => !v)}
      />
      {(task.subtasks?.length || addingSub) && (
        <div className="ml-6 border-l pl-3 pb-2">
          {task.subtasks?.map((sub) => (
            <TaskRowInner
              key={sub.id}
              projectId={projectId}
              task={sub}
              canEdit={canEdit}
              onChange={onChange}
              isSub
            />
          ))}
          {addingSub && (
            <AddTaskForm
              projectId={projectId}
              parentId={task.id}
              onChange={() => {
                setAddingSub(false);
                onChange();
              }}
            />
          )}
        </div>
      )}
    </div>
  );
}

function TaskRowInner({
  projectId,
  task,
  canEdit,
  onChange,
  onAddSub,
  isSub = false,
}: {
  projectId: string;
  task: Task;
  canEdit: boolean;
  onChange: () => void;
  onAddSub?: () => void;
  isSub?: boolean;
}) {
  async function changeStatus(status: TaskStatus) {
    try {
      await api.updateTask(projectId, task.id, { status });
      onChange();
    } catch {
      toast.error("Could not update task");
    }
  }

  async function remove() {
    try {
      await api.deleteTask(projectId, task.id);
      onChange();
    } catch {
      toast.error("Could not delete task");
    }
  }

  return (
    <div className="flex items-center gap-2 p-3">
      {isSub && <CornerDownRight className="h-4 w-4 shrink-0 text-muted-foreground" />}
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">{task.title}</p>
        {task.description && (
          <p className="truncate text-xs text-muted-foreground">
            {task.description}
          </p>
        )}
      </div>

      {canEdit ? (
        <Select value={task.status} onValueChange={(v) => changeStatus(v as TaskStatus)}>
          <SelectTrigger className="w-[140px] shrink-0">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {TASK_STATUSES.map((s) => (
              <SelectItem key={s} value={s}>
                {s}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      ) : (
        <span className="shrink-0 text-xs text-muted-foreground">{task.status}</span>
      )}

      {canEdit && !isSub && onAddSub && (
        <Button variant="ghost" size="icon" onClick={onAddSub} title="Add subtask">
          <Plus className="h-4 w-4" />
        </Button>
      )}
      {canEdit && (
        <Button variant="ghost" size="icon" onClick={remove} title="Delete task">
          <Trash2 className="h-4 w-4" />
        </Button>
      )}
    </div>
  );
}

function AddTaskForm({
  projectId,
  parentId,
  onChange,
}: {
  projectId: string;
  parentId?: string;
  onChange: () => void;
}) {
  const [title, setTitle] = React.useState("");
  const [submitting, setSubmitting] = React.useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!title.trim()) return;
    setSubmitting(true);
    try {
      await api.createTask(projectId, { title, parentId });
      setTitle("");
      onChange();
    } catch {
      toast.error("Could not add task");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={submit} className="flex gap-2 py-2">
      <Input
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder={parentId ? "New subtask…" : "Add a task…"}
      />
      <Button type="submit" size="sm" disabled={submitting}>
        Add
      </Button>
    </form>
  );
}
