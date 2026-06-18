"use client";

import * as React from "react";
import { toast } from "sonner";
import { Box, ExternalLink, Plus, Trash2 } from "lucide-react";

import { api } from "@/lib/api";
import {
  PRINT_SOURCES,
  type PrintLink,
  type PrintSource,
  type Task,
} from "@/lib/types";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
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
}

// flatten top-level tasks and their subtasks into one list.
function flatten(tasks: Task[]): Task[] {
  return tasks.flatMap((t) => [t, ...(t.subtasks ?? [])]);
}

export function PrintLinksPanel({ projectId, tasks, canEdit }: Props) {
  const [links, setLinks] = React.useState<PrintLink[] | null>(null);
  const flatTasks = React.useMemo(() => flatten(tasks), [tasks]);

  const load = React.useCallback(() => {
    api
      .listPrintLinks(projectId)
      .then(setLinks)
      .catch(() => toast.error("Could not load 3D files"));
  }, [projectId]);

  React.useEffect(() => load(), [load]);

  const taskTitle = (id: string) =>
    flatTasks.find((t) => t.id === id)?.title ?? "Unknown task";

  return (
    <div className="flex flex-col gap-4">
      {canEdit && (
        <div className="flex justify-end">
          <AddFileDialog projectId={projectId} tasks={flatTasks} onChange={load} />
        </div>
      )}

      {links === null ? (
        <p className="text-sm text-muted-foreground">Loading…</p>
      ) : links.length === 0 ? (
        <p className="text-sm text-muted-foreground">No 3D files yet.</p>
      ) : (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
          {links.map((link) => (
            <FileCard
              key={link.id}
              link={link}
              taskName={taskTitle(link.taskId)}
              canEdit={canEdit}
              onDelete={async () => {
                try {
                  await api.deletePrintLink(projectId, link.id);
                  load();
                } catch {
                  toast.error("Could not delete file");
                }
              }}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function FileCard({
  link,
  taskName,
  canEdit,
  onDelete,
}: {
  link: PrintLink;
  taskName: string;
  canEdit: boolean;
  onDelete: () => void;
}) {
  return (
    <div className="group relative overflow-hidden rounded-lg border">
      <a
        href={link.url}
        target="_blank"
        rel="noreferrer"
        className="block aspect-square bg-muted"
      >
        {link.thumbnailUrl ? (
          // eslint-disable-next-line @next/next/no-img-element -- user-supplied external thumbnails
          <img
            src={link.thumbnailUrl}
            alt={link.title || "3D model thumbnail"}
            className="h-full w-full object-cover"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center text-muted-foreground">
            <Box className="h-10 w-10" />
          </div>
        )}
      </a>
      <div className="flex flex-col gap-1 p-2">
        <a
          href={link.url}
          target="_blank"
          rel="noreferrer"
          className="flex items-center gap-1 truncate text-sm font-medium hover:underline"
        >
          {link.title || link.url}
          <ExternalLink className="h-3 w-3 shrink-0" />
        </a>
        <div className="flex items-center justify-between gap-1">
          <Badge variant="outline" className="truncate">
            {link.source}
          </Badge>
          <span className="truncate text-xs text-muted-foreground" title={taskName}>
            {taskName}
          </span>
        </div>
      </div>
      {canEdit && (
        <Button
          variant="destructive"
          size="icon"
          title="Delete file"
          className="absolute right-1 top-1 h-7 w-7 opacity-0 transition-opacity group-hover:opacity-100"
          onClick={onDelete}
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      )}
    </div>
  );
}

function AddFileDialog({
  projectId,
  tasks,
  onChange,
}: {
  projectId: string;
  tasks: Task[];
  onChange: () => void;
}) {
  const [open, setOpen] = React.useState(false);
  const [taskId, setTaskId] = React.useState("");
  const [source, setSource] = React.useState<PrintSource>("printables");
  const [url, setUrl] = React.useState("");
  const [thumbnailUrl, setThumbnailUrl] = React.useState("");
  const [title, setTitle] = React.useState("");
  const [submitting, setSubmitting] = React.useState(false);
  const [previewing, setPreviewing] = React.useState(false);

  // Resolve the page's OpenGraph thumbnail/title, filling empty fields only.
  async function fetchPreview() {
    if (!url) return;
    setPreviewing(true);
    try {
      const p = await api.previewLink(url);
      setThumbnailUrl((cur) => cur || p.thumbnailUrl);
      setTitle((cur) => cur || p.title);
    } catch {
      // best-effort: ignore preview failures
    } finally {
      setPreviewing(false);
    }
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!taskId) {
      toast.error("Pick a task for this file");
      return;
    }
    setSubmitting(true);
    try {
      await api.addPrintLink(projectId, { taskId, source, url, thumbnailUrl, title });
      toast.success("3D file added");
      setUrl("");
      setThumbnailUrl("");
      setTitle("");
      setOpen(false);
      onChange();
    } catch {
      toast.error("Could not add file");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button size="sm" disabled={tasks.length === 0}>
          <Plus className="h-4 w-4" />
          Add 3D file
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add 3D file</DialogTitle>
        </DialogHeader>
        {tasks.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            Add a task first — 3D files are attached to tasks.
          </p>
        ) : (
          <form onSubmit={submit} className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label>Task</Label>
              <Select value={taskId} onValueChange={setTaskId}>
                <SelectTrigger>
                  <SelectValue placeholder="Select a task" />
                </SelectTrigger>
                <SelectContent>
                  {tasks.map((t) => (
                    <SelectItem key={t.id} value={t.id}>
                      {t.title}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="f-source">Source</Label>
              <Select value={source} onValueChange={(v) => setSource(v as PrintSource)}>
                <SelectTrigger id="f-source">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {PRINT_SOURCES.map((s) => (
                    <SelectItem key={s} value={s}>
                      {s}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="f-url">Model URL</Label>
              <Input
                id="f-url"
                type="url"
                required
                placeholder="https://…"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                onBlur={fetchPreview}
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="f-thumb">
                Thumbnail URL{" "}
                <span className="font-normal text-muted-foreground">
                  {previewing ? "(fetching…)" : "(auto-detected, editable)"}
                </span>
              </Label>
              <Input
                id="f-thumb"
                type="url"
                placeholder="https://…/preview.png"
                value={thumbnailUrl}
                onChange={(e) => setThumbnailUrl(e.target.value)}
              />
              {thumbnailUrl && (
                // eslint-disable-next-line @next/next/no-img-element -- external preview
                <img
                  src={thumbnailUrl}
                  alt="thumbnail preview"
                  className="mt-1 h-24 w-24 rounded border object-cover"
                />
              )}
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="f-title">Title (optional)</Label>
              <Input
                id="f-title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
              />
            </div>
            <DialogFooter>
              <Button type="submit" disabled={submitting}>
                {submitting ? "Adding…" : "Add file"}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
}
