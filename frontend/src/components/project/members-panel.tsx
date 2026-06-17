"use client";

import * as React from "react";
import { toast } from "sonner";
import { Trash2, UserPlus } from "lucide-react";

import { api, ApiError } from "@/lib/api";
import type { Membership, Role } from "@/lib/types";
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

const ROLES: Role[] = ["admin", "editor", "viewer"];

interface Props {
  projectId: string;
  members: Membership[];
  ownerId: string;
  isAdmin: boolean;
  onChange: () => void;
}

export function MembersPanel({ projectId, members, ownerId, isAdmin, onChange }: Props) {
  return (
    <div className="flex flex-col gap-3">
      {isAdmin && <AddMemberForm projectId={projectId} onChange={onChange} />}
      <ul className="flex flex-col gap-2">
        {members.map((m) => (
          <li
            key={m.id}
            className="flex items-center justify-between gap-2 rounded-md border px-3 py-2"
          >
            <div className="min-w-0">
              <p className="truncate text-sm font-medium">
                {m.user?.name || m.user?.email || m.userId}
              </p>
              {m.user?.email && (
                <p className="truncate text-xs text-muted-foreground">
                  {m.user.email}
                </p>
              )}
            </div>
            <div className="flex items-center gap-2">
              <Badge variant="secondary">{m.role}</Badge>
              {isAdmin && m.userId !== ownerId && (
                <Button
                  variant="ghost"
                  size="icon"
                  title="Remove member"
                  onClick={async () => {
                    try {
                      await api.removeMember(projectId, m.userId);
                      onChange();
                    } catch {
                      toast.error("Could not remove member");
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
    </div>
  );
}

function AddMemberForm({
  projectId,
  onChange,
}: {
  projectId: string;
  onChange: () => void;
}) {
  const [email, setEmail] = React.useState("");
  const [role, setRole] = React.useState<Role>("viewer");
  const [submitting, setSubmitting] = React.useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    try {
      await api.addMember(projectId, email, role);
      toast.success("Member added");
      setEmail("");
      onChange();
    } catch (err) {
      toast.error(
        err instanceof ApiError && err.status === 404
          ? "No user with that email has signed in yet"
          : "Could not add member"
      );
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={submit} className="flex flex-col gap-2 sm:flex-row">
      <Input
        type="email"
        required
        placeholder="member@example.com"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
      />
      <Select value={role} onValueChange={(v) => setRole(v as Role)}>
        <SelectTrigger className="sm:w-[130px]">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {ROLES.map((r) => (
            <SelectItem key={r} value={r}>
              {r}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Button type="submit" disabled={submitting}>
        <UserPlus className="h-4 w-4" />
        Add
      </Button>
    </form>
  );
}
