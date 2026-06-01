import { useEffect, useState } from "react";
import { Check, Pencil, ShieldAlert, X } from "lucide-react";
import { useParams } from "react-router-dom";

import { AgentInsight } from "../../components/agent/AgentInsight";
import { Button } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import { EmptyState } from "../../components/ui/EmptyState";
import {
  correctLongTermMemoryFact,
  getCustomerDetail,
  getCustomerLongTermMemory,
  rejectLongTermMemoryFact
} from "../../lib/api/customers";
import { mockTasks } from "../../lib/mock/dashboard";
import type { CustomerDetail } from "../../types/customerDetail";
import type { LongTermMemory, LongTermMemoryFact } from "../../types/memory";
import { useAuth } from "../auth/use-auth";

interface CorrectionDraft {
  category: string;
  key: string;
  value: string;
  reason: string;
}

function createDraft(fact: LongTermMemoryFact): CorrectionDraft {
  return {
    category: fact.category,
    key: fact.key,
    value: fact.value,
    reason: "Human correction from memory manager"
  };
}

export function CustomerDetailPage() {
  const { customerId = "cus_001" } = useParams();
  const { user } = useAuth();
  const [detail, setDetail] = useState<CustomerDetail | null>(null);
  const [memory, setMemory] = useState<LongTermMemory | null>(null);
  const [errorText, setErrorText] = useState("");
  const [memoryErrorText, setMemoryErrorText] = useState("");
  const [actionText, setActionText] = useState("");
  const [editingFactId, setEditingFactId] = useState("");
  const [draft, setDraft] = useState<CorrectionDraft | null>(null);
  const task = mockTasks[0];
  const recommendation = detail?.latestRecommendation ?? task.recommendation;
  const customerName = detail?.customer.name ?? memory?.customer.name ?? task.customer.name;

  async function loadMemory(active = true) {
    try {
      const result = await getCustomerLongTermMemory(customerId);
      if (active) {
        setMemory(result);
        setMemoryErrorText("");
      }
    } catch {
      if (active) {
        setMemory(null);
        setMemoryErrorText("长期记忆暂时不可用，请确认后端服务和数据库连接状态。");
      }
    }
  }

  useEffect(() => {
    let active = true;

    void getCustomerDetail(customerId)
      .then((detailResult) => {
        if (active) {
          setDetail(detailResult);
          setErrorText("");
        }
      })
      .catch(() => {
        if (active) {
          setDetail(null);
          setErrorText("当前身份无权查看该客户，或后端服务暂未启动。");
        }
      });

    void getCustomerLongTermMemory(customerId)
      .then((memoryResult) => {
        if (active) {
          setMemory(memoryResult);
          setMemoryErrorText("");
        }
      })
      .catch(() => {
        if (active) {
          setMemory(null);
          setMemoryErrorText("长期记忆暂时不可用，请确认后端服务和数据库连接状态。");
        }
      });

    return () => {
      active = false;
    };
  }, [customerId, user.id, user.role]);

  async function rejectFact(fact: LongTermMemoryFact) {
    setActionText("正在拒绝长期记忆事实...");
    try {
      await rejectLongTermMemoryFact(customerId, fact.id, "Rejected from customer memory manager");
      await loadMemory();
      setActionText("已拒绝该事实，后续 Agent 提示词不会再引用它。");
    } catch {
      setActionText("拒绝失败，请稍后重试或检查后端日志。");
    }
  }

  function beginCorrection(fact: LongTermMemoryFact) {
    setEditingFactId(fact.id);
    setDraft(createDraft(fact));
    setActionText("");
  }

  async function submitCorrection(fact: LongTermMemoryFact) {
    if (!draft || !draft.category.trim() || !draft.key.trim() || !draft.value.trim()) {
      setActionText("修正内容不能为空。");
      return;
    }

    setActionText("正在保存修正...");
    try {
      await correctLongTermMemoryFact(customerId, fact.id, {
        category: draft.category.trim(),
        key: draft.key.trim(),
        value: draft.value.trim(),
        confidence: Math.max(fact.confidence, 0.95),
        reason: draft.reason.trim() || "Human correction from memory manager"
      });
      await loadMemory();
      setEditingFactId("");
      setDraft(null);
      setActionText("已保存修正，旧事实已归档，新事实会进入 Agent 长期记忆。");
    } catch {
      setActionText("修正失败，请稍后重试或检查后端日志。");
    }
  }

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">{customerName}</h1>
          <p className="page__subtitle">客户画像、聊天记录、长期记忆和 AI 推荐在这里统一复盘。</p>
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
            {detail.profileEvidence.map((item) => (
              <span key={item}>{item}</span>
            ))}
          </div>
        </Card>
      ) : null}

      <Card className="memory-manager">
        <div className="memory-manager__header">
          <div>
            <h2 className="state-panel__title">长期记忆管理</h2>
            <p className="memory-manager__hint">只展示 active 事实。拒绝或修正后，后续 Agent 提示词会自动避开旧事实。</p>
          </div>
          {memory ? <span className="memory-manager__status">构建时间：{memory.builtAt}</span> : null}
        </div>

        {memoryErrorText ? <p className="memory-manager__status">{memoryErrorText}</p> : null}
        {actionText ? <p className="memory-manager__status">{actionText}</p> : null}

        {memory && memory.facts.length > 0 ? (
          <div className="memory-fact-list">
            {memory.facts.map((fact) => (
              <article className="memory-fact" key={fact.id}>
                <div className="memory-fact__main">
                  <div className="memory-fact__meta">
                    <strong>{fact.category}.{fact.key}</strong>
                    <span>置信度 {Math.round(fact.confidence * 100)}%</span>
                    <span>来源 {fact.sourceType}</span>
                  </div>
                  <p className="memory-fact__value">{fact.value}</p>
                </div>

                {editingFactId === fact.id && draft ? (
                  <div className="memory-correction-form">
                    <label className="form-field">
                      <span className="form-field__label">分类</span>
                      <input
                        className="form-field__input"
                        value={draft.category}
                        onChange={(event) => setDraft({ ...draft, category: event.target.value })}
                      />
                    </label>
                    <label className="form-field">
                      <span className="form-field__label">键名</span>
                      <input className="form-field__input" value={draft.key} onChange={(event) => setDraft({ ...draft, key: event.target.value })} />
                    </label>
                    <label className="form-field memory-correction-form__wide">
                      <span className="form-field__label">事实内容</span>
                      <textarea
                        className="form-field__textarea"
                        value={draft.value}
                        onChange={(event) => setDraft({ ...draft, value: event.target.value })}
                      />
                    </label>
                    <label className="form-field memory-correction-form__wide">
                      <span className="form-field__label">修正原因</span>
                      <input
                        className="form-field__input"
                        value={draft.reason}
                        onChange={(event) => setDraft({ ...draft, reason: event.target.value })}
                      />
                    </label>
                    <div className="memory-fact__actions memory-correction-form__wide">
                      <Button type="button" variant="primary" onClick={() => void submitCorrection(fact)}>
                        <Check size={16} aria-hidden="true" />
                        保存
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={() => {
                          setEditingFactId("");
                          setDraft(null);
                        }}
                      >
                        <X size={16} aria-hidden="true" />
                        取消
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className="memory-fact__actions">
                    <Button type="button" variant="secondary" onClick={() => beginCorrection(fact)}>
                      <Pencil size={16} aria-hidden="true" />
                      修正
                    </Button>
                    <Button type="button" variant="ghost" onClick={() => void rejectFact(fact)}>
                      <ShieldAlert size={16} aria-hidden="true" />
                      拒绝
                    </Button>
                  </div>
                )}
              </article>
            ))}
          </div>
        ) : null}

        {memory && memory.facts.length === 0 ? (
          <div className="state-panel">
            <h2 className="state-panel__title">暂无长期记忆</h2>
            <p className="state-panel__message">当前客户还没有可用于提示词的 active 事实。</p>
          </div>
        ) : null}
      </Card>
    </section>
  );
}
