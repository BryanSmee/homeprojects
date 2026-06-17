"use client";

import { useParams } from "next/navigation";

import { useAuth } from "@/components/auth-provider";
import { ProjectView } from "@/components/project/project-view";

export default function PublicProjectPage() {
  const { user, loading } = useAuth();
  const params = useParams<{ id: string }>();

  if (loading) return null;
  return <ProjectView projectId={params.id} currentUserId={user?.id ?? null} />;
}
