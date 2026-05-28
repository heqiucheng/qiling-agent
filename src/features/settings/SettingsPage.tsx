import { EmptyState } from "../../components/ui/EmptyState";

export function SettingsPage() {
  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">设置</h1>
          <p className="page__subtitle">MVP 仅保留必要配置入口。</p>
        </div>
      </header>
      <EmptyState title="设置骨架已就绪" message="后续加入个人信息、角色信息、企业微信配置入口和 AI 分析偏好。" />
    </section>
  );
}