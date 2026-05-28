import { AgentInsight } from "../../components/agent/AgentInsight";
import { ScriptTaskCard } from "../../components/followup/ScriptTaskCard";
import { MetricCard } from "../../components/ui/MetricCard";
import { mockInsights, mockMetrics, mockTasks } from "../../lib/mock/dashboard";

export function DashboardPage() {
  const firstTask = mockTasks[0];
  const firstInsight = mockInsights[0];

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">工作台</h1>
          <p className="page__subtitle">今天有 12 条待确认话术，3 个高意向客户需要优先跟进。</p>
        </div>
      </header>
      <div className="page-grid">
        <div className="page-grid page-grid--two">
          {mockMetrics.map((metric) => <MetricCard key={metric.label} {...metric} />)}
        </div>
        <div className="page-grid page-grid--two">
          <ScriptTaskCard task={firstTask} onCopy={() => undefined} />
          <AgentInsight action={firstTask.recommendation.recommendedAction} reasoning={firstInsight.evidence} riskFlags={[firstInsight.suggestion]} />
        </div>
      </div>
    </section>
  );
}
