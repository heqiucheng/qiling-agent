import { Card } from "./Card";

interface EmptyStateProps {
  title: string;
  message: string;
  action?: string;
}

export function EmptyState({ title, message, action }: EmptyStateProps) {
  return (
    <Card className="state-panel">
      <h2 className="state-panel__title">{title}</h2>
      <p className="state-panel__message">{message}</p>
      {action ? <span>{action}</span> : null}
    </Card>
  );
}