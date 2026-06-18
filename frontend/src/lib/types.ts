export type Role = "admin" | "editor" | "viewer";

export type TaskStatus = "Waiting" | "In Progress" | "Done" | "Abandoned";
export type ProjectStatus = TaskStatus;

export const TASK_STATUSES: TaskStatus[] = [
  "Waiting",
  "In Progress",
  "Done",
  "Abandoned",
];

export const PRINT_SOURCES = [
  "thingiverse",
  "printables",
  "cults3d",
  "makerworld",
  "other",
] as const;
export type PrintSource = (typeof PRINT_SOURCES)[number];

export interface User {
  id: string;
  email: string;
  name: string;
  createdAt: string;
  updatedAt: string;
}

export interface Task {
  id: string;
  projectId: string;
  parentId?: string | null;
  title: string;
  description: string;
  status: TaskStatus;
  createdAt: string;
  updatedAt: string;
  subtasks?: Task[];
}

export interface Membership {
  id: string;
  projectId: string;
  userId: string;
  role: Role;
  createdAt: string;
  updatedAt: string;
  user?: User;
}

export interface Project {
  id: string;
  name: string;
  description: string;
  ownerId: string;
  public: boolean;
  status: ProjectStatus;
  createdAt: string;
  updatedAt: string;
}

export interface ProjectDetail extends Project {
  tasks: Task[];
  members?: Membership[];
}

export interface PrintLink {
  id: string;
  projectId: string;
  taskId: string;
  source: PrintSource;
  url: string;
  thumbnailUrl: string;
  title: string;
  notes: string;
  status?: string;
  createdAt: string;
  updatedAt: string;
}

export interface AuthConfig {
  oidcEnabled: boolean;
}
