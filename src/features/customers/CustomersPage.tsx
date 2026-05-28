import { EmptyState } from "../../components/ui/EmptyState";

export function CustomersPage() {
  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">客户</h1>
          <p className="page__subtitle">查看客户阶段、意向、主要顾虑和待跟进状态。</p>
        </div>
      </header>
      <EmptyState title="客户列表骨架已就绪" message="下一步会接入 Mock 客户表格和筛选器。" action="可先从数据接入页导入聊天记录" />
    </section>
  );
}