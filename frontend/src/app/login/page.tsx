"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { KeyRound } from "lucide-react";

import { useAuth } from "@/components/auth-provider";
import { api, ApiError } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function LoginPage() {
  const { user, loading, refresh } = useAuth();
  const router = useRouter();
  const [email, setEmail] = React.useState("");
  const [name, setName] = React.useState("");
  const [submitting, setSubmitting] = React.useState(false);
  const [oidcEnabled, setOidcEnabled] = React.useState<boolean | null>(null);

  React.useEffect(() => {
    api
      .authConfig()
      .then((c) => setOidcEnabled(c.oidcEnabled))
      .catch(() => setOidcEnabled(false));
  }, []);

  React.useEffect(() => {
    if (!loading && user) router.replace("/");
  }, [loading, user, router]);

  async function handleDevLogin(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    try {
      await api.devLogin(email, name);
      await refresh();
      router.push("/");
    } catch (err) {
      const msg =
        err instanceof ApiError && err.status === 404
          ? "Dev login is disabled (SSO is configured). Use Sign in with SSO."
          : "Login failed";
      toast.error(msg);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="mx-auto flex max-w-md flex-col gap-6 px-4 py-16">
      <Card>
        <CardHeader>
          <CardTitle>Welcome back</CardTitle>
          <CardDescription>Sign in to manage your home projects.</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-6">
          {oidcEnabled === null ? (
            <p className="text-center text-sm text-muted-foreground">Loading…</p>
          ) : oidcEnabled ? (
            <Button asChild className="w-full">
              <a href={api.loginUrl()}>
                <KeyRound className="h-4 w-4" />
                Sign in with SSO
              </a>
            </Button>
          ) : (
            <form onSubmit={handleDevLogin} className="flex flex-col gap-3">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="you@example.com"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Your name"
                />
              </div>
              <Button type="submit" disabled={submitting}>
                {submitting ? "Signing in…" : "Developer login"}
              </Button>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
