"use client";

import * as React from "react";
import { useParams, useRouter } from "next/navigation";

import { useAuth } from "@/components/auth-provider";
import { ProjectView } from "@/components/project/project-view";

export default function ProjectPage() {
  const { user, loading } = useAuth();
  const router = useRouter();
  const params = useParams<{ id: string }>();

  React.useEffect(() => {
    if (!loading && !user) router.replace("/login");
  }, [loading, user, router]);

  if (loading || !user) return null;
  return <ProjectView projectId={params.id} currentUserId={user.id} />;
}
