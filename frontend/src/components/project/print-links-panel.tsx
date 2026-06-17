"use client";

import * as React from "react";
import { toast } from "sonner";
import { ExternalLink, Trash2, Box } from "lucide-react";

import { api } from "@/lib/api";
import { PRINT_SOURCES, type PrintLink, type PrintSource } from "@/lib/types";
import { Badge } from "@/components/ui/badge";
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
  canEdit: boolean;
}

export function PrintLinksPanel({ projectId, canEdit }: Props) {
  const [links, setLinks] = React.useState<PrintLink[] | null>(null);

  const load = React.useCallback(() => {
    api
      .listPrintLinks(projectId)
      .then(setLinks)
      .catch(() => toast.error("Could not load print links"));
  }, [projectId]);

  React.useEffect(() => load(), [load]);

  return (
    <div className="flex flex-col gap-3">
      {canEdit && <AddLinkForm projectId={projectId} onChange={load} />}
      {links && links.length === 0 ? (
        <p className="text-sm text-muted-foreground">No 3D model links yet.</p>
      ) : (
        <ul className="flex flex-col gap-2">
          {links?.map((link) => (
            <li
              key={link.id}
              className="flex items-center justify-between gap-2 rounded-md border px-3 py-2"
            >
              <div className="flex min-w-0 items-center gap-2">
                <Box className="h-4 w-4 shrink-0 text-muted-foreground" />
                <div className="min-w-0">
                  <a
                    href={link.url}
                    target="_blank"
                    rel="noreferrer"
                    className="flex items-center gap-1 truncate text-sm font-medium hover:underline"
                  >
                    {link.title || link.url}
                    <ExternalLink className="h-3 w-3 shrink-0" />
                  </a>
                  {link.notes && (
                    <p className="truncate text-xs text-muted-foreground">
                      {link.notes}
                    </p>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Badge variant="outline">{link.source}</Badge>
                {canEdit && (
                  <Button
                    variant="ghost"
                    size="icon"
                    title="Delete link"
                    onClick={async () => {
                      try {
                        await api.deletePrintLink(projectId, link.id);
                        load();
                      } catch {
                        toast.error("Could not delete link");
                      }
                    }}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                )}
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function AddLinkForm({
  projectId,
  onChange,
}: {
  projectId: string;
  onChange: () => void;
}) {
  const [url, setUrl] = React.useState("");
  const [title, setTitle] = React.useState("");
  const [source, setSource] = React.useState<PrintSource>("printables");
  const [submitting, setSubmitting] = React.useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    try {
      await api.addPrintLink(projectId, { url, title, source });
      toast.success("Link added");
      setUrl("");
      setTitle("");
      onChange();
    } catch {
      toast.error("Could not add link");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={submit} className="flex flex-col gap-2">
      <div className="flex flex-col gap-2 sm:flex-row">
        <Input
          required
          type="url"
          placeholder="https://…"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
        />
        <Select value={source} onValueChange={(v) => setSource(v as PrintSource)}>
          <SelectTrigger className="sm:w-[150px]">
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
      <div className="flex gap-2">
        <Input
          placeholder="Title (optional)"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
        />
        <Button type="submit" disabled={submitting}>
          Add link
        </Button>
      </div>
    </form>
  );
}
