import { useEffect, useState } from "react";

import { AgentInsight } from "../../components/agent/AgentInsight";
import { MetricCard } from "../../components/ui/MetricCard";
import { useAuth } from "../auth/use-auth";
import { getReviewSummary } from "../../lib/api/review";
import { mockInsights, mockMetrics } from "../../lib/mock/dashboard";
import type { ReviewSummary } from "../../types/review";

export function ReviewCenterPage() {
  const { user } = useAuth();
  const [summary, setSummary] = useState<ReviewSummary | null>(null);
  const insights = summary?.insights ?? mockInsights;
  const metrics = summary?.metrics ?? mockMetrics;
  const insight = insights[0];

  useEffect(() => {
    let active = true;

    void getReviewSummary()
      .then((result) => {
        if (active) {
          setSummary(result);
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

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">复盘中心</h1>
          <p className="page__subtitle">查看机会、风险、话术效果和改进建议。</p>
          {summary?.sampleWarning ? <p className="page__subtitle">{summary.sampleWarning}</p> : null}
        </div>
      </header>
      <div className="page-grid">
        <div className="page-grid page-grid--two">
          {metrics.slice(0, 4).map((metric) => <MetricCard key={metric.label} {...metric} />)}
        </div>
        <AgentInsight action={insight.title} reasoning={insight.evidence} riskFlags={[insight.suggestion]} />
      </div>
    </section>
  );
}
