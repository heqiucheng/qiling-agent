import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { AgentInsight } from "../../components/agent/AgentInsight";
import { Card } from "../../components/ui/Card";
import { EmptyState } from "../../components/ui/EmptyState";
import { useAuth } from "../auth/use-auth";
import { getCustomerDetail } from "../../lib/api/customers";
import { mockTasks } from "../../lib/mock/dashboard";
import type { CustomerDetail } from "../../types/customerDetail";

export function CustomerDetailPage() {
  const { customerId = "cus_001" } = useParams();
  const { user } = useAuth();
  const [detail, setDetail] = useState<CustomerDetail | null>(null);
  const [errorText, setErrorText] = useState("");
  const task = mockTasks[0];
  const recommendation = detail?.latestRecommendation ?? task.recommendation;
  const customerName = detail?.customer.name ?? task.customer.name;

  useEffect(() => {
    let active = true;

    void getCustomerDetail(customerId)
      .then((result) => {
        if (active) {
          setDetail(result);
          setErrorText("");
        }
      })
      .catch(() => {
        if (active) {
          setDetail(null);
          setErrorText("当前身份无权查看该客户，或后端服务未启动。");
        }
      });

    return () => {
      active = false;
    };
  }, [customerId, user.id, user.role]);

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">{customerName}</h1>
          <p className="page__subtitle">客户画像、聊天记录和 AI 推荐将在这里汇总。</p>
        </div>
      </header>
      <div className="page-grid page-grid--two">
        {detail ? (
          <Card className="timeline">
            <h2 className="state-panel__title">聊天记录时间线</h2>
            <div className="timeline__list">
              {detail.conversationMessages.map((message) => (
                <div className="timeline__item" key={message.id}>
                  <strong>{message.senderName}</strong>
                  <span>{message.sentAt}</span>
                  <p>{message.content}</p>
                </div>
              ))}
            </div>
          </Card>
        ) : (
          <EmptyState title="聊天记录时间线" message={errorText || "后端未启动时显示占位状态。"} />
        )}
        <AgentInsight action={recommendation.recommendedAction} reasoning={recommendation.reasoning} riskFlags={recommendation.riskFlags} />
      </div>
      {detail ? (
        <Card className="result-panel">
          <strong>客户画像依据</strong>
          <div className="result-panel__grid">
            {detail.profileEvidence.map((item) => <span key={item}>{item}</span>)}
          </div>
        </Card>
      ) : null}
    </section>
  );
}
