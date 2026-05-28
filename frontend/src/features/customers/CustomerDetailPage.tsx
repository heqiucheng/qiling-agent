import { AgentInsight } from "../../components/agent/AgentInsight";
import { EmptyState } from "../../components/ui/EmptyState";
import { mockTasks } from "../../lib/mock/dashboard";

export function CustomerDetailPage() {
  const task = mockTasks[0];

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">{task.customer.name}</h1>
          <p className="page__subtitle">客户画像、聊天记录和 AI 推荐将在这里汇总。</p>
        </div>
      </header>
      <div className="page-grid page-grid--two">
        <EmptyState title="聊天记录时间线" message="下一步接入上传解析后的消息列表。" />
        <AgentInsight action={task.recommendation.recommendedAction} reasoning={task.recommendation.reasoning} riskFlags={task.recommendation.riskFlags} />
      </div>
    </section>
  );
}