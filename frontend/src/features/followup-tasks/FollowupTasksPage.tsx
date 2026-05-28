import { ScriptTaskCard } from "../../components/followup/ScriptTaskCard";
import { mockTasks } from "../../lib/mock/dashboard";

export function FollowupTasksPage() {
  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">待确认话术</h1>
          <p className="page__subtitle">集中处理 Agent 自动生成的待确认话术。</p>
        </div>
      </header>
      <div className="page-grid">
        {mockTasks.map((task) => <ScriptTaskCard key={task.id} task={task} onCopy={() => undefined} />)}
      </div>
    </section>
  );
}