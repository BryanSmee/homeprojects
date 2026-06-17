"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Globe, Lock, Pencil, Share2, Trash2 } from "lucide-react";

import { api, ApiError } from "@/lib/api";
import type { ProjectDetail, Role } from "@/lib/types";
import { StatusBadge } from "@/components/status-badge";
import { TaskList } from "@/components/project/task-list";
import { MembersPanel } from "@/components/project/members-panel";
import { PrintLinksPanel } from "@/components/project/print-links-panel";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

function roleOf(project: ProjectDetail, userId: string | null): Role | null {
  if (!userId) return null;
  if (project.ownerId === userId) return "admin";
  return project.members?.find((m) => m.userId === userId)?.role ?? null;
}

export function ProjectView({
  projectId,
  currentUserId,
}: {
  projectId: string;
  currentUserId: string | null;
}) {
  const [project, setProject] = React.useState<ProjectDetail | null>(null);
  const [error, setError] = React.useState<string | null>(null);

  const load = React.useCallback(() => {
    api
      .getProject(projectId)
      .then(setProject)
      .catch((err) => {
        if (err instanceof ApiError && (err.status === 401 || err.status === 403))
          setError("This project is private.");
        else if (err instanceof ApiError && err.status === 404)
          setError("Project not found.");
        else setError("Could not load project.");
      });
  }, [projectId]);

  React.useEffect(() => load(), [load]);

  if (error)
    return (
      <div className="mx-auto max-w-2xl px-4 py-16 text-center text-muted-foreground">
        {error}
      </div>
    );
  if (!project) return null;

  const role = roleOf(project, currentUserId);
  const canEdit = role === "admin" || role === "editor";
  const isAdmin = role === "admin";

  return (
    <div className="mx-auto flex max-w-3xl flex-col gap-6 px-4 py-8">
      <ProjectHeader
        project={project}
        isAdmin={isAdmin}
        onChange={load}
      />

      <Card>
        <CardHeader>
          <CardTitle>Tasks</CardTitle>
        </CardHeader>
        <CardContent>
          <TaskList
            projectId={projectId}
            tasks={project.tasks ?? []}
            canEdit={canEdit}
            onChange={load}
          />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>3D printing</CardTitle>
        </CardHeader>
        <CardContent>
          <PrintLinksPanel projectId={projectId} canEdit={canEdit} />
        </CardContent>
      </Card>

      {project.members && (
        <Card>
          <CardHeader>
            <CardTitle>Members</CardTitle>
          </CardHeader>
          <CardContent>
            <MembersPanel
              projectId={projectId}
              members={project.members}
              ownerId={project.ownerId}
              isAdmin={isAdmin}
              onChange={load}
            />
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function ProjectHeader({
  project,
  isAdmin,
  onChange,
}: {
  project: ProjectDetail;
  isAdmin: boolean;
  onChange: () => void;
}) {
  const router = useRouter();

  async function toggleVisibility() {
    try {
      await api.setVisibility(project.id, !project.public);
      onChange();
    } catch {
      toast.error("Could not change visibility");
    }
  }

  async function remove() {
    if (!confirm("Delete this project and all its tasks?")) return;
    try {
      await api.deleteProject(project.id);
      toast.success("Project deleted");
      router.push("/");
    } catch {
      toast.error("Could not delete project");
    }
  }

  function copyShareLink() {
    const url = `${window.location.origin}/p/${project.id}`;
    navigator.clipboard.writeText(url);
    toast.success("Public link copied");
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-3">
            <h1 className="truncate text-2xl font-semibold">{project.name}</h1>
            <StatusBadge status={project.status} />
          </div>
          {project.description && (
            <p className="mt-1 text-muted-foreground">{project.description}</p>
          )}
        </div>
      </div>

      {isAdmin && (
        <div className="flex flex-wrap items-center gap-2">
          <EditProjectDialog project={project} onChange={onChange} />
          <Button variant="outline" size="sm" onClick={toggleVisibility}>
            {project.public ? (
              <>
                <Lock className="h-4 w-4" /> Make private
              </>
            ) : (
              <>
                <Globe className="h-4 w-4" /> Make public
              </>
            )}
          </Button>
          {project.public && (
            <Button variant="outline" size="sm" onClick={copyShareLink}>
              <Share2 className="h-4 w-4" /> Copy link
            </Button>
          )}
          <Button variant="destructive" size="sm" onClick={remove}>
            <Trash2 className="h-4 w-4" /> Delete
          </Button>
        </div>
      )}
    </div>
  );
}

function EditProjectDialog({
  project,
  onChange,
}: {
  project: ProjectDetail;
  onChange: () => void;
}) {
  const [open, setOpen] = React.useState(false);
  const [name, setName] = React.useState(project.name);
  const [description, setDescription] = React.useState(project.description);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    try {
      await api.updateProject(project.id, name, description);
      toast.success("Project updated");
      setOpen(false);
      onChange();
    } catch {
      toast.error("Could not update project");
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm">
          <Pencil className="h-4 w-4" /> Edit
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit project</DialogTitle>
        </DialogHeader>
        <form onSubmit={submit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="e-name">Name</Label>
            <Input
              id="e-name"
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="e-desc">Description</Label>
            <Textarea
              id="e-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button type="submit">Save</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
