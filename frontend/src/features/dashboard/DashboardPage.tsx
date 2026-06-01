import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { AgentInsight } from "../../components/agent/AgentInsight";
import { ScriptTaskCard } from "../../components/followup/ScriptTaskCard";
import { MetricCard } from "../../components/ui/MetricCard";
import { getDashboardSummary } from "../../lib/api/dashboard";
import { mockInsights, mockMetrics, mockTasks } from "../../lib/mock/dashboard";
import type { DashboardSummary } from "../../types/dashboard";
import { useAuth } from "../auth/use-auth";

export function DashboardPage() {
  const { user } = useAuth();
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [isFallback, setIsFallback] = useState(false);

  useEffect(() => {
    let active = true;

    void getDashboardSummary()
      .then((data) => {
        if (active) {
          setSummary(data);
          setIsFallback(false);
        }
      })
      .catch(() => {
        if (active) {
          setSummary(null);
          setIsFallback(true);
        }
      });

    return () => {
      active = false;
    };
  }, [user.id, user.role]);

  const metrics = summary?.metrics ?? mockMetrics;
  const firstTask = summary?.priorityTasks[0] ?? mockTasks[0];
  const firstInsight = summary?.dailyReview
    ? { title: "今日复盘", evidence: summary.dailyReview.summary, suggestion: summary.dailyReview.suggestions[0] ?? "继续跟进高意向客户。" }
    : mockInsights[0];
  const subtitle = summary?.dailyReview.summary ?? "今日优先处理高意向客户、沉默客户和待确认话术。";

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">工作台</h1>
          <p className="page__subtitle">{subtitle}</p>
          {isFallback ? <p className="page__subtitle">当前后端不可用，页面正在展示兜底演示数据。</p> : null}
        </div>
        <div className="task-card__actions">
          <Link className="button button--primary button-link" to="/app/data-ingestion">
            接入聊天记录
          </Link>
          <Link className="button button--secondary button-link" to="/app/followup-tasks">
            处理待确认话术
          </Link>
        </div>
      </header>
      <div className="page-grid">
        <div className="page-grid page-grid--two">
          {metrics.map((metric) => (
            <MetricCard key={metric.label} {...metric} />
          ))}
        </div>
        <div className="page-grid page-grid--two">
          <ScriptTaskCard task={firstTask} onCopy={() => undefined} />
          <AgentInsight action={firstInsight.title} reasoning={firstInsight.evidence} riskFlags={[firstInsight.suggestion]} />
        </div>
      </div>
    </section>
  );
}
