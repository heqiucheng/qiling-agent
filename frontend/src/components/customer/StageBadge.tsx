import { Badge } from "../ui/Badge";
import { stageLabels } from "../../lib/format";
import type { CustomerStage } from "../../types/customer";

const stageTone: Record<CustomerStage, string> = {
  new_lead: "stage-new",
  opened: "stage-new",
  needs_discovery: "stage-new",
  product_interested: "stage-new",
  price_objection: "stage-objection",
  high_intent: "stage-commit",
  closing: "stage-commit",
  won: "stage-commit",
  silent: "stage-silent",
  churn_risk: "stage-risk"
};

interface StageBadgeProps {
  stage: CustomerStage;
}

export function StageBadge({ stage }: StageBadgeProps) {
  return <Badge tone={stageTone[stage]}>{stageLabels[stage]}</Badge>;
}