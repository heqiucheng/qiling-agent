import { useMemo, useState, type ReactNode } from "react";

import { AuthContext, type AuthContextValue } from "./auth-context-value";
import type { MockRole, MockUser } from "./auth-types";

const users: Record<MockRole, MockUser> = {
  sales: { id: "usr_001", name: "销售A", role: "sales", roleLabel: "销售视角" },
  manager: { id: "mgr_001", name: "主管", role: "manager", roleLabel: "主管视角" }
};

export function AuthProvider({ children }: { children: ReactNode }) {
  const [role, setRole] = useState<MockRole>(() => {
    const stored = window.localStorage.getItem("qiling_mock_role");
    return stored === "manager" ? "manager" : "sales";
  });

  const value = useMemo<AuthContextValue>(
    () => ({
      user: users[role],
      switchRole: (nextRole) => {
        window.localStorage.setItem("qiling_mock_role", nextRole);
        setRole(nextRole);
      }
    }),
    [role]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
