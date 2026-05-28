import { useEffect, useState } from "react";

import { ScriptTaskCard } from "../../components/followup/ScriptTaskCard";
import { EmptyState } from "../../components/ui/EmptyState";
import { copyFollowupTask, listPendingFollowupTasks, markFollowupTaskWrong, regenerateFollowupTask, skipFollowupTask } from "../../lib/api/followupTasks";
import { mockTasks } from "../../lib/mock/dashboard";
import type { FollowupTask } from "../../types/followup";

export function FollowupTasksPage() {
  const [tasks, setTasks] = useState<FollowupTask[]>(mockTasks);
  const [statusText, setStatusText] = useState("等待处理待确认话术");
  const [busyTaskId, setBusyTaskId] = useState<string | null>(null);

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

  async function handleCopy(task: FollowupTask) {
    setBusyTaskId(task.id);
    try {
      await navigator.clipboard.writeText(task.recommendation.script);
      await copyFollowupTask(task.id, task.recommendation.script);
      setTasks((current) => current.filter((item) => item.id !== task.id));
      setStatusText(`已复制 ${task.customer.name} 的话术`);
    } catch (error) {
      setStatusText(error instanceof Error ? error.message : "复制话术失败");
    } finally {
      setBusyTaskId(null);
    }
  }

  async function handleRegenerate(taskId: string) {
    setBusyTaskId(taskId);
    try {
      const recommendation = await regenerateFollowupTask(taskId);
      setTasks((current) => current.map((task) => (task.id === taskId ? { ...task, recommendation } : task)));
      setStatusText("已换一种话术");
    } catch (error) {
      setStatusText(error instanceof Error ? error.message : "换一种话术失败");
    } finally {
      setBusyTaskId(null);
    }
  }

  async function handleSkip(taskId: string) {
    setBusyTaskId(taskId);
    try {
      await skipFollowupTask(taskId);
      setTasks((current) => current.filter((task) => task.id !== taskId));
      setStatusText("已跳过该任务");
    } catch (error) {
      setStatusText(error instanceof Error ? error.message : "跳过任务失败");
    } finally {
      setBusyTaskId(null);
    }
  }

  async function handleMarkWrong(taskId: string) {
    setBusyTaskId(taskId);
    try {
      await markFollowupTaskWrong(taskId);
      setTasks((current) => current.filter((task) => task.id !== taskId));
      setStatusText("已记录不准反馈");
    } catch (error) {
      setStatusText(error instanceof Error ? error.message : "标记不准失败");
    } finally {
      setBusyTaskId(null);
    }
  }

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">待确认话术</h1>
          <p className="page__subtitle">集中处理 Agent 自动生成的待确认话术。</p>
          <p className="page__subtitle">{statusText}</p>
        </div>
      </header>
      {tasks.length > 0 ? (
        <div className="page-grid">
          {tasks.map((task) => (
            <ScriptTaskCard
              key={task.id}
              task={task}
              onCopy={() => void handleCopy(task)}
              onRegenerate={handleRegenerate}
              onSkip={handleSkip}
              onMarkWrong={handleMarkWrong}
              isBusy={busyTaskId === task.id}
            />
          ))}
        </div>
      ) : (
        <EmptyState title="暂无待确认话术" message="上传聊天记录后，Agent 会自动生成待确认任务。" action="去数据接入页上传聊天记录" />
      )}
    </section>
  );
}
