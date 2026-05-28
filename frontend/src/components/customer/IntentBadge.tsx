import { Badge } from "../ui/Badge";
import { intentLabels } from "../../lib/format";
import type { IntentLevel } from "../../types/customer";

interface IntentBadgeProps {
  level: IntentLevel;
}

export function IntentBadge({ level }: IntentBadgeProps) {
  return <Badge tone={`intent-${level}`}>{intentLabels[level]}</Badge>;
}