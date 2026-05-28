import { useEffect, useState } from "react";

import { AgentInsight } from "../../components/agent/AgentInsight";
import { ScriptTaskCard } from "../../components/followup/ScriptTaskCard";
import { MetricCard } from "../../components/ui/MetricCard";
import { useAuth } from "../auth/use-auth";
import { getDashboardSummary } from "../../lib/api/dashboard";
import { mockInsights, mockMetrics, mockTasks } from "../../lib/mock/dashboard";
import type { DashboardSummary } from "../../types/dashboard";

export function DashboardPage() {
  const { user } = useAuth();
  const [summary, setSummary] = useState<DashboardSummary | null>(null);

  useEffect(() => {
    let active = true;

    void getDashboardSummary()
      .then((data) => {
        if (active) {
          setSummary(data);
        }
      })
      .catch(() => {
        if (active) {
          setSummary(null);
        }
      });

    return () => {
      active = false;
    };
  }, [user.id, user.role]);

  const metrics = summary?.metrics ?? mockMetrics;
  const firstTask = summary?.priorityTasks[0] ?? mockTasks[0];
  const firstInsight = mockInsights[0];
  const subtitle = summary?.dailyReview.summary ?? "今天有 12 条待确认话术，3 个高意向客户需要优先跟进。";

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">工作台</h1>
          <p className="page__subtitle">{subtitle}</p>
        </div>
      </header>
      <div className="page-grid">
        <div className="page-grid page-grid--two">
          {metrics.map((metric) => <MetricCard key={metric.label} {...metric} />)}
        </div>
        <div className="page-grid page-grid--two">
          <ScriptTaskCard task={firstTask} onCopy={() => undefined} />
          <AgentInsight action={firstTask.recommendation.recommendedAction} reasoning={firstInsight.evidence} riskFlags={[firstInsight.suggestion]} />
        </div>
      </div>
    </section>
  );
}
