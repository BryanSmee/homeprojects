import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { TaskStatus } from "@/lib/types";

const styles: Record<TaskStatus, string> = {
  Waiting: "bg-muted text-muted-foreground",
  "In Progress": "bg-blue-600 text-white",
  Done: "bg-green-600 text-white",
  Abandoned: "bg-zinc-400 text-white line-through",
};

export function StatusBadge({
  status,
  className,
}: {
  status: TaskStatus;
  className?: string;
}) {
  return (
    <Badge className={cn("border-transparent", styles[status], className)}>
      {status}
    </Badge>
  );
}
