import { Navigate, createBrowserRouter } from "react-router-dom";
import { AppShell } from "../components/layout/AppShell";
import { LoginPage } from "../features/auth/LoginPage";
import { DashboardPage } from "../features/dashboard/DashboardPage";
import { CustomerDetailPage } from "../features/customers/CustomerDetailPage";
import { CustomersPage } from "../features/customers/CustomersPage";
import { DataIngestionPage } from "../features/data-ingestion/DataIngestionPage";
import { FollowupTasksPage } from "../features/followup-tasks/FollowupTasksPage";
import { ReportCenterPage } from "../features/reports/ReportCenterPage";
import { ReviewCenterPage } from "../features/review-center/ReviewCenterPage";
import { SettingsPage } from "../features/settings/SettingsPage";

export const router = createBrowserRouter([
  { path: "/", element: <Navigate to="/app/dashboard" replace /> },
  { path: "/login", element: <LoginPage /> },
  {
    path: "/app",
    element: <AppShell />,
    children: [
      { index: true, element: <Navigate to="/app/dashboard" replace /> },
      { path: "dashboard", element: <DashboardPage /> },
      { path: "customers", element: <CustomersPage /> },
      { path: "customers/:customerId", element: <CustomerDetailPage /> },
      { path: "followup-tasks", element: <FollowupTasksPage /> },
      { path: "data-ingestion", element: <DataIngestionPage /> },
      { path: "reports", element: <ReportCenterPage /> },
      { path: "review-center", element: <ReviewCenterPage /> },
      { path: "settings", element: <SettingsPage /> }
    ]
  }
]);
