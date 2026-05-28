import { EmptyState } from "../../components/ui/EmptyState";

export function DataIngestionPage() {
  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">数据接入</h1>
          <p className="page__subtitle">连接企业微信，或先上传聊天记录体验 Agent。</p>
        </div>
      </header>
      <div className="page-grid page-grid--two">
        <EmptyState title="企业微信接入" message="后续接入前必须先完成本地接口验证和 Mock。" action="当前保留配置入口" />
        <EmptyState title="上传/粘贴聊天记录" message="MVP 主链路会优先实现该入口。" action="支持 txt/csv 优先" />
      </div>
    </section>
  );
}