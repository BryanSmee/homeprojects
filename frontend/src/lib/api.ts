import type {
  AuthConfig,
  Membership,
  PrintLink,
  PrintSource,
  Project,
  ProjectDetail,
  Role,
  Task,
  TaskStatus,
  User,
} from "./types";

export const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    credentials: "include",
    headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) },
  });

  if (!res.ok) {
    const message = await res
      .json()
      .then((b) => b?.error as string)
      .catch(() => null);
    throw new ApiError(res.status, message ?? `Request failed (${res.status})`);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

const body = (data: unknown) => JSON.stringify(data);

export const api = {
  // Auth
  authConfig: () => request<AuthConfig>("/api/auth/config"),
  me: () => request<User>("/api/auth/me"),
  devLogin: (email: string, name: string) =>
    request<User>("/api/auth/dev-login", {
      method: "POST",
      body: body({ email, name }),
    }),
  logout: () => request<void>("/api/auth/logout", { method: "POST" }),
  loginUrl: () => `${API_BASE}/api/auth/login`,

  // Projects
  listProjects: () => request<Project[]>("/api/projects"),
  createProject: (name: string, description: string) =>
    request<Project>("/api/projects", {
      method: "POST",
      body: body({ name, description }),
    }),
  getProject: (id: string) => request<ProjectDetail>(`/api/projects/${id}`),
  updateProject: (id: string, name: string, description: string) =>
    request<Project>(`/api/projects/${id}`, {
      method: "PATCH",
      body: body({ name, description }),
    }),
  deleteProject: (id: string) =>
    request<void>(`/api/projects/${id}`, { method: "DELETE" }),
  setVisibility: (id: string, isPublic: boolean) =>
    request<Project>(`/api/projects/${id}/visibility`, {
      method: "PATCH",
      body: body({ public: isPublic }),
    }),

  // Members
  listMembers: (projectId: string) =>
    request<Membership[]>(`/api/projects/${projectId}/members`),
  addMember: (projectId: string, email: string, role: Role) =>
    request<Membership>(`/api/projects/${projectId}/members`, {
      method: "POST",
      body: body({ email, role }),
    }),
  removeMember: (projectId: string, userId: string) =>
    request<void>(`/api/projects/${projectId}/members/${userId}`, {
      method: "DELETE",
    }),

  // Tasks
  createTask: (
    projectId: string,
    input: { title: string; description?: string; parentId?: string | null }
  ) =>
    request<Task>(`/api/projects/${projectId}/tasks`, {
      method: "POST",
      body: body(input),
    }),
  updateTask: (
    projectId: string,
    taskId: string,
    input: { title?: string; description?: string; status?: TaskStatus }
  ) =>
    request<Task>(`/api/projects/${projectId}/tasks/${taskId}`, {
      method: "PATCH",
      body: body(input),
    }),
  deleteTask: (projectId: string, taskId: string) =>
    request<void>(`/api/projects/${projectId}/tasks/${taskId}`, {
      method: "DELETE",
    }),

  // Printing extension
  previewLink: (url: string) =>
    request<{ title: string; thumbnailUrl: string }>(
      `/api/ext/printing/preview?url=${encodeURIComponent(url)}`
    ),
  listPrintLinks: (projectId: string, taskId?: string) =>
    request<PrintLink[]>(
      `/api/ext/printing/projects/${projectId}/links` +
        (taskId ? `?taskId=${encodeURIComponent(taskId)}` : "")
    ),
  addPrintLink: (
    projectId: string,
    input: {
      taskId: string;
      source: PrintSource;
      url: string;
      thumbnailUrl?: string;
      title?: string;
      notes?: string;
    }
  ) =>
    request<PrintLink>(`/api/ext/printing/projects/${projectId}/links`, {
      method: "POST",
      body: body(input),
    }),
  deletePrintLink: (projectId: string, linkId: string) =>
    request<void>(`/api/ext/printing/projects/${projectId}/links/${linkId}`, {
      method: "DELETE",
    }),
};
