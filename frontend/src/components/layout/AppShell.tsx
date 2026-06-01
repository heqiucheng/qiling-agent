import { BarChart3, DatabaseZap, FileText, Home, MessageSquareText, Settings, Users } from "lucide-react";
import { NavLink, Outlet } from "react-router-dom";

import { useAuth } from "../../features/auth/use-auth";

const navItems = [
  { to: "/app/dashboard", label: "工作台", icon: Home },
  { to: "/app/customers", label: "客户", icon: Users },
  { to: "/app/followup-tasks", label: "待确认话术", icon: MessageSquareText },
  { to: "/app/data-ingestion", label: "数据接入", icon: DatabaseZap },
  { to: "/app/reports", label: "报告中心", icon: FileText },
  { to: "/app/review-center", label: "复盘中心", icon: BarChart3 },
  { to: "/app/settings", label: "设置", icon: Settings }
];

export function AppShell() {
  const { user, switchRole } = useAuth();
  const nextRoleLabel = user.role === "sales" ? "主管视角" : "销售视角";

  return (
    <div className="app-shell">
      <aside className="side-nav" aria-label="主导航">
        <div className="side-nav__brand">
          <span className="side-nav__name">企灵 Agent</span>
          <span className="side-nav__subtitle">私域客户分析、跟进与复盘</span>
        </div>
        <nav className="side-nav__items">
          {navItems.map((item) => {
            const Icon = item.icon;
            return (
              <NavLink key={item.to} to={item.to} className={({ isActive }) => `side-nav__item${isActive ? " active" : ""}`}>
                <Icon size={18} aria-hidden="true" />
                <span>{item.label}</span>
              </NavLink>
            );
          })}
        </nav>
      </aside>
      <main className="app-main">
        <header className="top-bar">
          <div className="top-bar__meta">
            <span>演示企业</span>
            <span>{user.roleLabel}</span>
            <span>{user.name}</span>
          </div>
          <div className="top-bar__meta">
            <span>同步状态：连接本地后端，失败时使用兜底数据</span>
            <button className="role-switch" type="button" onClick={() => switchRole(user.role === "sales" ? "manager" : "sales")}>
              切换为{nextRoleLabel}
            </button>
          </div>
        </header>
        <Outlet />
      </main>
    </div>
  );
}
