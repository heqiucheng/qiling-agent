import type { Report } from "../../types/report";
import { apiPost } from "./client";
import type { ReportDto } from "./dto";
import { mapReport } from "./mappers";

export async function generateCustomerIntentReport(range = "last_7_days"): Promise<Report> {
  const data = await apiPost<ReportDto, Record<string, unknown>>("/api/reports/customer-intent", { range });
  return mapReport(data);
}
