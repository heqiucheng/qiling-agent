import { useNavigate } from "react-router-dom";

import { Button } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import type { MockRole } from "./auth-types";
import { useAuth } from "./use-auth";

export function LoginPage() {
  const navigate = useNavigate();
  const { switchRole } = useAuth();

  function enter(role: MockRole) {
    switchRole(role);
    navigate("/app/dashboard", { replace: true });
  }

  return (
    <main className="login-page">
      <Card className="login-panel">
        <div>
          <h1 className="page__title">企灵 Agent</h1>
          <p className="page__subtitle">选择一个演示身份进入系统。</p>
        </div>
        <div className="task-card__actions">
          <Button variant="primary" onClick={() => enter("sales")}>销售视角</Button>
          <Button variant="secondary" onClick={() => enter("manager")}>主管视角</Button>
        </div>
      </Card>
    </main>
  );
}
