import { Link } from "react-router-dom";
import { useState } from "react";

import { Button } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import { EmptyState } from "../../components/ui/EmptyState";
import { confirmUpload, uploadConversation } from "../../lib/api/uploads";
import type { ConfirmUploadResult, UploadConversationResult } from "../../types/upload";

const sampleConversation = "王女士 09:20 这个价格还能优惠吗？\n销售A 09:22 我先结合您的需求整理一个方案。";

export function DataIngestionPage() {
  const [content, setContent] = useState(sampleConversation);
  const [uploadResult, setUploadResult] = useState<UploadConversationResult | null>(null);
  const [confirmResult, setConfirmResult] = useState<ConfirmUploadResult | null>(null);
  const [statusText, setStatusText] = useState("等待上传聊天记录");
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleUpload() {
    if (!content.trim()) {
      setStatusText("聊天记录不能为空。");
      return;
    }

    setIsSubmitting(true);
    setConfirmResult(null);
    try {
      const result = await uploadConversation(content);
      setUploadResult(result);
      setStatusText("解析完成，等待确认客户归属。");
    } catch (error) {
      setStatusText(error instanceof Error ? error.message : "上传解析失败");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleConfirm() {
    if (!uploadResult) {
      return;
    }

    setIsSubmitting(true);
    try {
      const result = await confirmUpload(uploadResult.uploadId, uploadResult.parsedCustomer.name);
      setConfirmResult(result);
      setStatusText("已生成客户画像、Agent 运行记录和待确认话术任务。");
    } catch (error) {
      setStatusText(error instanceof Error ? error.message : "确认解析结果失败");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">数据接入</h1>
          <p className="page__subtitle">先支持粘贴聊天记录，确认后自动生成客户画像、Agent Run 和待确认任务。</p>
        </div>
      </header>
      <div className="page-grid page-grid--two">
        <EmptyState title="企业微信接入" message="正式对接前，必须先完成本地接口验证、鉴权、回调签名和 Mock 演练。" action="当前保留配置入口" />
        <Card className="form-panel">
          <div>
            <h2 className="state-panel__title">上传/粘贴聊天记录</h2>
            <p className="state-panel__message">MVP 优先支持粘贴文本，确认后进入客户详情和待确认话术流程。</p>
          </div>
          <label className="form-field">
            <span className="form-field__label">聊天记录</span>
            <textarea className="form-field__textarea" value={content} onChange={(event) => setContent(event.target.value)} rows={8} />
          </label>
          <div className="task-card__actions">
            <Button variant="primary" onClick={handleUpload} disabled={isSubmitting}>
              上传解析
            </Button>
            <Button variant="secondary" onClick={handleConfirm} disabled={!uploadResult || isSubmitting}>
              确认并生成任务
            </Button>
          </div>
          <div className="result-panel" aria-live="polite">
            <strong>{statusText}</strong>
            {uploadResult ? (
              <div className="result-panel__grid">
                <span>上传 ID：{uploadResult.uploadId}</span>
                <span>客户：{uploadResult.parsedCustomer.name}</span>
                <span>负责人：{uploadResult.parsedCustomer.ownerName}</span>
                <span>消息数：{uploadResult.messageCount}</span>
              </div>
            ) : null}
            {confirmResult ? (
              <>
                <div className="result-panel__grid">
                  <span>客户 ID：{confirmResult.customerId}</span>
                  <span>任务 ID：{confirmResult.followupTaskId}</span>
                  <span>Agent Run：{confirmResult.agentRunId}</span>
                </div>
                <div className="task-card__actions">
                  <Link className="button button--primary button-link" to={`/app/customers/${confirmResult.customerId}`}>
                    查看客户详情
                  </Link>
                  <Link className="button button--secondary button-link" to="/app/followup-tasks">
                    处理待确认话术
                  </Link>
                </div>
              </>
            ) : null}
          </div>
        </Card>
      </div>
    </section>
  );
}
