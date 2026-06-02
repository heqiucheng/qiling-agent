import type { PageResult } from "../../types/api";
import type { Report, ReportSummary } from "../../types/report";
import { apiGet, apiPost } from "./client";
import type { PageDto, ReportDto, ReportSummaryDto } from "./dto";
import { mapPage, mapReport, mapReportSummary } from "./mappers";

export async function generateCustomerIntentReport(range = "last_7_days"): Promise<Report> {
  const data = await apiPost<ReportDto, Record<string, unknown>>("/api/reports/customer-intent", { range });
  return mapReport(data);
}

export async function getReports(): Promise<PageResult<ReportSummary>> {
  const data = await apiGet<PageDto<ReportSummaryDto>>("/api/reports");
  return mapPage(data, mapReportSummary);
}

export async function getReport(reportId: string): Promise<Report> {
  const data = await apiGet<ReportDto>(`/api/reports/${reportId}`);
  return mapReport(data);
}
