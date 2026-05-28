import { AgentInsight } from "../../components/agent/AgentInsight";
import { MetricCard } from "../../components/ui/MetricCard";
import { mockInsights, mockMetrics } from "../../lib/mock/dashboard";

export function ReviewCenterPage() {
  const insight = mockInsights[0];

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">复盘中心</h1>
          <p className="page__subtitle">查看机会、风险、话术效果和改进建议。</p>
        </div>
      </header>
      <div className="page-grid">
        <div className="page-grid page-grid--two">
          {mockMetrics.slice(0, 2).map((metric) => <MetricCard key={metric.label} {...metric} />)}
        </div>
        <AgentInsight action={insight.title} reasoning={insight.evidence} riskFlags={[insight.suggestion]} />
      </div>
    </section>
  );
}