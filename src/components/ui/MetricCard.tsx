import { Card } from "./Card";

interface MetricCardProps {
  label: string;
  value: string;
  hint: string;
}

export function MetricCard({ label, value, hint }: MetricCardProps) {
  return (
    <Card className="metric-card">
      <span className="metric-card__label">{label}</span>
      <strong className="metric-card__value">{value}</strong>
      <span className="metric-card__hint">{hint}</span>
    </Card>
  );
}