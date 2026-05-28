import { Button } from "../ui/Button";
import { Card } from "../ui/Card";
import { IntentBadge } from "../customer/IntentBadge";
import { StageBadge } from "../customer/StageBadge";
import type { FollowupTask } from "../../types/followup";

interface ScriptTaskCardProps {
  task: FollowupTask;
  onCopy: (script: string) => void;
  onRegenerate?: (taskId: string) => void;
  onSkip?: (taskId: string) => void;
  onMarkWrong?: (taskId: string) => void;
  isBusy?: boolean;
}

export function ScriptTaskCard({ task, onCopy, onRegenerate, onSkip, onMarkWrong, isBusy = false }: ScriptTaskCardProps) {
  const { customer, recommendation } = task;

  return (
    <Card className="task-card">
      <div className="task-card__meta">
        <strong>{customer.name}</strong>
        <StageBadge stage={recommendation.customerStage} />
        <IntentBadge level={recommendation.intentLevel} />
        <span>{task.type}</span>
      </div>
      <p className="task-card__script">{recommendation.script}</p>
      <div className="agent-insight__section">
        <span className="agent-insight__label">推荐理由</span>
        <p className="agent-insight__text">{recommendation.reasoning}</p>
      </div>
      <div className="task-card__actions">
        <Button variant="primary" onClick={() => onCopy(recommendation.script)} disabled={isBusy}>复制话术</Button>
        <Button variant="secondary" onClick={() => onRegenerate?.(task.id)} disabled={isBusy}>换一种</Button>
        <Button variant="ghost" onClick={() => onSkip?.(task.id)} disabled={isBusy}>跳过</Button>
        <Button variant="ghost" onClick={() => onMarkWrong?.(task.id)} disabled={isBusy}>标记不准</Button>
      </div>
    </Card>
  );
}
