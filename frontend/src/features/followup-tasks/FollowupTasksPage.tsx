import { useEffect, useState } from "react";

import { ScriptTaskCard } from "../../components/followup/ScriptTaskCard";
import { EmptyState } from "../../components/ui/EmptyState";
import { listPendingFollowupTasks } from "../../lib/api/followupTasks";
import { mockTasks } from "../../lib/mock/dashboard";
import type { FollowupTask } from "../../types/followup";

export function FollowupTasksPage() {
  const [tasks, setTasks] = useState<FollowupTask[]>(mockTasks);

  useEffect(() => {
    let active = true;

    void listPendingFollowupTasks()
      .then((result) => {
        if (active) {
          setTasks(result.items);
        }
      })
      .catch(() => {
        if (active) {
          setTasks(mockTasks);
        }
      });

    return () => {
      active = false;
    };
  }, []);

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">待确认话术</h1>
          <p className="page__subtitle">集中处理 Agent 自动生成的待确认话术。</p>
        </div>
      </header>
      {tasks.length > 0 ? (
        <div className="page-grid">
          {tasks.map((task) => <ScriptTaskCard key={task.id} task={task} onCopy={() => undefined} />)}
        </div>
      ) : (
        <EmptyState title="暂无待确认话术" message="上传聊天记录后，Agent 会自动生成待确认任务。" action="去数据接入页上传聊天记录" />
      )}
    </section>
  );
}
