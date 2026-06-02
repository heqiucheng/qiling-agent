import type { PageResult } from "../../types/api";
import type { Report, ReportExportTask, ReportSummary } from "../../types/report";
import { apiGet, apiGetBlob, apiGetText, apiPost } from "./client";
import type { PageDto, ReportDto, ReportExportTaskDto, ReportSummaryDto } from "./dto";
import { mapPage, mapReport, mapReportExportTask, mapReportSummary } from "./mappers";

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

export async function exportReportMarkdown(reportId: string): Promise<string> {
  return apiGetText(`/api/reports/${reportId}/export?format=markdown`);
}

export async function exportReportXLSX(reportId: string): Promise<Blob> {
  return apiGetBlob(`/api/reports/${reportId}/export?format=xlsx`);
}

export async function exportReportDOCX(reportId: string): Promise<Blob> {
  return apiGetBlob(`/api/reports/${reportId}/export?format=docx`);
}

export async function exportReportPDF(reportId: string): Promise<Blob> {
  return apiGetBlob(`/api/reports/${reportId}/export?format=pdf`);
}

export async function createReportExportTask(reportId: string, format = "pdf"): Promise<ReportExportTask> {
  const data = await apiPost<ReportExportTaskDto, Record<string, unknown>>(`/api/reports/${reportId}/export-tasks`, { format });
  return mapReportExportTask(data);
}

export async function getReportExportTasks(): Promise<PageResult<ReportExportTask>> {
  const data = await apiGet<PageDto<ReportExportTaskDto>>("/api/report-export-tasks");
  return mapPage(data, mapReportExportTask);
}

export async function getReportExportTask(taskId: string): Promise<ReportExportTask> {
  const data = await apiGet<ReportExportTaskDto>(`/api/report-export-tasks/${taskId}`);
  return mapReportExportTask(data);
}
