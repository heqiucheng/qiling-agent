import { Bot } from "lucide-react";

import { Card } from "../ui/Card";

interface AgentInsightProps {
  action: string;
  reasoning: string;
  riskFlags: string[];
}

export function AgentInsight({ action, reasoning, riskFlags }: AgentInsightProps) {
  return (
    <Card className="agent-insight">
      <div className="agent-insight__title">
        <Bot size={18} aria-hidden="true" />
        AI 建议
      </div>
      <div className="agent-insight__section">
        <span className="agent-insight__label">推荐动作</span>
        <p className="agent-insight__text">{action}</p>
      </div>
      <div className="agent-insight__section">
        <span className="agent-insight__label">推荐理由</span>
        <p className="agent-insight__text">{reasoning}</p>
      </div>
      <div className="agent-insight__section">
        <span className="agent-insight__label">风险提示</span>
        <p className="agent-insight__text">{riskFlags.join("，")}</p>
      </div>
    </Card>
  );
}
