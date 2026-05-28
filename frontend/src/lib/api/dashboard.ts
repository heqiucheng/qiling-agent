import type { DashboardSummary } from "../../types/dashboard";
import { apiGet } from "./client";
import type { DashboardSummaryDto } from "./dto";
import { mapDashboardSummary } from "./mappers";

export async function getDashboardSummary(): Promise<DashboardSummary> {
  const data = await apiGet<DashboardSummaryDto>("/api/dashboard/summary");
  return mapDashboardSummary(data);
}
