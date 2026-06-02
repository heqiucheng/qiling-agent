import { ClipboardList, Copy, Download, FileSpreadsheet, FileText, FileType, RefreshCw } from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import { MetricCard } from "../../components/ui/MetricCard";
import { createReportExportTask, exportReportDOCX, exportReportMarkdown, exportReportPDF, exportReportXLSX, generateCustomerIntentReport, getReport, getReportExportTasks, getReports } from "../../lib/api/reports";
import { copyText } from "../../lib/clipboard";
import { downloadBlob, downloadTextFile } from "../../lib/download";
import type { Report, ReportExportTask, ReportSummary } from "../../types/report";
import { useAuth } from "../auth/use-auth";

export function ReportCenterPage() {
  const { user } = useAuth();
  const [report, setReport] = useState<Report | null>(null);
  const [history, setHistory] = useState<ReportSummary[]>([]);
  const [exportTasks, setExportTasks] = useState<ReportExportTask[]>([]);
  const [latestExportTask, setLatestExportTask] = useState<ReportExportTask | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [exportingFormat, setExportingFormat] = useState<"markdown" | "xlsx" | "docx" | "pdf" | null>(null);
  const [isCreatingExportTask, setIsCreatingExportTask] = useState(false);
  const [isHistoryLoading, setIsHistoryLoading] = useState(true);
  const [status, setStatus] = useState("");

  async function refreshHistory() {
    const result = await getReports();
    setHistory(result.items);
    return result.items;
  }

  async function refreshExportTasks() {
    const result = await getReportExportTasks();
    setExportTasks(result.items);
    setLatestExportTask(result.items[0] ?? null);
    return result.items;
  }

  async function generateReport() {
    setIsLoading(true);
    setStatus("");
    try {
      const result = await generateCustomerIntentReport();
      setReport(result);
      await refreshHistory();
      setStatus("报告已生成");
    } catch {
      setStatus("报告生成失败，请确认后端服务可用");
    } finally {
      setIsLoading(false);
    }
  }

  async function openReport(reportId: string) {
    setIsHistoryLoading(true);
    setStatus("");
    try {
      const result = await getReport(reportId);
      setReport(result);
      setStatus("已加载历史报告");
    } catch {
      setStatus("历史报告加载失败，请稍后重试");
    } finally {
      setIsHistoryLoading(false);
    }
  }

  async function copyMarkdown() {
    if (!report) {
      return;
    }
    const copied = await copyText(report.markdown);
    setStatus(copied ? "Markdown 已复制" : "当前浏览器不支持自动复制");
  }

  async function downloadMarkdown() {
    if (!report) {
      return;
    }
    setExportingFormat("markdown");
    setStatus("");
    try {
      const markdown = await exportReportMarkdown(report.id);
      downloadTextFile(`${report.title}-${report.id}.md`, markdown, "text/markdown;charset=utf-8");
      setStatus("Markdown 已下载");
    } catch {
      setStatus("Markdown 下载失败，请稍后重试");
    } finally {
      setExportingFormat(null);
    }
  }

  async function downloadExcel() {
    if (!report) {
      return;
    }
    setExportingFormat("xlsx");
    setStatus("");
    try {
      const workbook = await exportReportXLSX(report.id);
      downloadBlob(`${report.title}-${report.id}.xlsx`, workbook);
      setStatus("Excel 报表已下载");
    } catch {
      setStatus("Excel 报表下载失败，请稍后重试");
    } finally {
      setExportingFormat(null);
    }
  }

  async function downloadWord() {
    if (!report) {
      return;
    }
    setExportingFormat("docx");
    setStatus("");
    try {
      const document = await exportReportDOCX(report.id);
      downloadBlob(`${report.title}-${report.id}.docx`, document);
      setStatus("Word 报告已下载");
    } catch {
      setStatus("Word 报告下载失败，请稍后重试");
    } finally {
      setExportingFormat(null);
    }
  }

  async function downloadPDF() {
    if (!report) {
      return;
    }
    setExportingFormat("pdf");
    setStatus("");
    try {
      const pdf = await exportReportPDF(report.id);
      downloadBlob(`${report.title}-${report.id}.pdf`, pdf);
      setStatus("PDF 报告已下载");
    } catch {
      setStatus("PDF 报告下载失败，请稍后重试");
    } finally {
      setExportingFormat(null);
    }
  }

  async function createPDFExportTask() {
    if (!report) {
      return;
    }
    setIsCreatingExportTask(true);
    setStatus("");
    try {
      const task = await createReportExportTask(report.id, "pdf");
      setLatestExportTask(task);
      await refreshExportTasks();
      setStatus(task.status === "completed" ? "PDF 导出任务已完成" : "PDF 导出任务失败，请查看任务记录");
    } catch {
      setStatus("PDF 导出任务创建失败，请稍后重试");
    } finally {
      setIsCreatingExportTask(false);
    }
  }

  useEffect(() => {
    let active = true;

    void Promise.all([getReports(), getReportExportTasks()])
      .then(async (result) => {
        if (!active) {
          return;
        }
        const [reportsResult, exportTasksResult] = result;
        setHistory(reportsResult.items);
        setExportTasks(exportTasksResult.items);
        setLatestExportTask(exportTasksResult.items[0] ?? null);
        if (reportsResult.items[0]) {
          const latest = await getReport(reportsResult.items[0].id);
          if (active) {
            setReport(latest);
          }
          return;
        }
        const generated = await generateCustomerIntentReport();
        if (active) {
          setReport(generated);
          await refreshHistory();
          await refreshExportTasks();
          setStatus("报告已生成");
        }
      })
      .catch(() => {
        if (active) {
          setStatus("报告生成失败，请确认后端服务可用");
        }
      })
      .finally(() => {
        if (active) {
          setIsHistoryLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [user.id, user.role]);

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">报告中心</h1>
          <p className="page__subtitle">生成客户意愿、风险、待确认事项和下一步行动清单。</p>
          {status ? <p className="page__subtitle">{status}</p> : null}
        </div>
        <div className="task-card__actions">
          <Button variant="secondary" onClick={generateReport} disabled={isLoading}>
            <RefreshCw size={16} aria-hidden="true" />
            {isLoading ? "生成中" : "重新生成"}
          </Button>
          <Button variant="primary" onClick={copyMarkdown} disabled={!report}>
            <Copy size={16} aria-hidden="true" />
            复制 Markdown
          </Button>
          <Button variant="secondary" onClick={downloadMarkdown} disabled={!report || exportingFormat !== null}>
            <Download size={16} aria-hidden="true" />
            {exportingFormat === "markdown" ? "下载中" : "下载 Markdown"}
          </Button>
          <Button variant="secondary" onClick={downloadExcel} disabled={!report || exportingFormat !== null}>
            <FileSpreadsheet size={16} aria-hidden="true" />
            {exportingFormat === "xlsx" ? "下载中" : "下载 Excel"}
          </Button>
          <Button variant="secondary" onClick={downloadWord} disabled={!report || exportingFormat !== null}>
            <FileType size={16} aria-hidden="true" />
            {exportingFormat === "docx" ? "下载中" : "下载 Word"}
          </Button>
          <Button variant="secondary" onClick={downloadPDF} disabled={!report || exportingFormat !== null}>
            <FileText size={16} aria-hidden="true" />
            {exportingFormat === "pdf" ? "下载中" : "下载 PDF"}
          </Button>
          <Button variant="secondary" onClick={createPDFExportTask} disabled={!report || isCreatingExportTask}>
            <ClipboardList size={16} aria-hidden="true" />
            {isCreatingExportTask ? "任务生成中" : "创建 PDF 任务"}
          </Button>
        </div>
      </header>

      {report ? (
        <div className="page-grid">
          <Card className="report-hero">
            <div>
              <div className="report-hero__eyebrow">
                <FileText size={16} aria-hidden="true" />
                {report.rangeLabel}
              </div>
              <h2>{report.title}</h2>
              <p>{report.summary}</p>
            </div>
            <span>{new Date(report.generatedAt).toLocaleString()}</span>
          </Card>

          <div className="page-grid page-grid--two">
            {report.metrics.map((metric) => (
              <MetricCard key={metric.label} {...metric} />
            ))}
          </div>

          <div className="report-layout">
            <div className="report-layout__main">
              {report.sections.map((section) => (
                <Card key={section.title} className="report-section">
                  <div>
                    <h2>{section.title}</h2>
                    <p>{section.summary}</p>
                  </div>
                  <div className="report-customer-list">
                    {section.items.map((item) => (
                      <article key={item.customerId} className="report-customer">
                        <div className="report-customer__header">
                          <strong>{item.customerName}</strong>
                          <span>{item.intent}</span>
                          <span>{item.stage}</span>
                        </div>
                        <p>{item.reasoning}</p>
                        <div className="task-card__script">{item.script}</div>
                        <div className="report-evidence">
                          {item.evidence.map((evidence) => (
                            <span key={evidence}>{evidence}</span>
                          ))}
                        </div>
                      </article>
                    ))}
                  </div>
                </Card>
              ))}
            </div>

            <aside className="report-layout__side">
              <Card className="report-history">
                <div className="report-history__header">
                  <h2>历史报告</h2>
                  <span>{isHistoryLoading ? "加载中" : `${history.length} 份`}</span>
                </div>
                <div className="report-history__list">
                  {history.length > 0 ? (
                    history.map((item) => (
                      <button
                        key={item.id}
                        className={`report-history__item${report.id === item.id ? " active" : ""}`}
                        type="button"
                        onClick={() => void openReport(item.id)}
                        disabled={isHistoryLoading}
                      >
                        <strong>{item.title}</strong>
                        <span>{new Date(item.generatedAt).toLocaleString()}</span>
                        <small>
                          {item.actionItemCount} 项行动 / {item.sectionCount} 个板块
                        </small>
                      </button>
                    ))
                  ) : (
                    <p>暂无历史报告</p>
                  )}
                </div>
              </Card>

              <Card className="report-actions">
                <h2>行动清单</h2>
                <div className="report-action-list">
                  {report.actionItems.map((item) => (
                    <article key={`${item.customerId}-${item.priority}`} className="report-action">
                      <span>{item.priority}</span>
                      <strong>{item.customerName}</strong>
                      <p>{item.action}</p>
                      <small>{item.dueHint}</small>
                    </article>
                  ))}
                </div>
              </Card>

              <Card className="report-export-tasks">
                <div className="report-history__header">
                  <h2>导出任务</h2>
                  <span>{exportTasks.length} 条</span>
                </div>
                {latestExportTask ? (
                  <div className="report-export-task">
                    <span className={`report-export-task__status report-export-task__status--${latestExportTask.status}`}>
                      {reportExportTaskStatusLabel(latestExportTask.status)}
                    </span>
                    <strong>{latestExportTask.filename || `${latestExportTask.format.toUpperCase()} 导出任务`}</strong>
                    <small>
                      {latestExportTask.format.toUpperCase()} / {formatBytes(latestExportTask.sizeBytes)} / {new Date(latestExportTask.createdAt).toLocaleString()}
                    </small>
                    {latestExportTask.error ? <p>{latestExportTask.error}</p> : null}
                  </div>
                ) : (
                  <p className="report-export-tasks__empty">暂无导出任务</p>
                )}
                <div className="report-export-task-list">
                  {exportTasks.slice(0, 5).map((task) => (
                    <div key={task.id} className="report-export-task-list__item">
                      <span>{reportExportTaskStatusLabel(task.status)}</span>
                      <strong>{task.filename || task.id}</strong>
                      <small>{new Date(task.createdAt).toLocaleString()}</small>
                    </div>
                  ))}
                </div>
              </Card>
            </aside>
          </div>
        </div>
      ) : (
        <Card className="report-empty">
          <h2>暂无报告</h2>
          <p>点击重新生成后，系统会基于当前可见客户、任务和 Agent 建议生成报告。</p>
        </Card>
      )}
    </section>
  );
}

function reportExportTaskStatusLabel(status: ReportExportTask["status"]): string {
  switch (status) {
    case "completed":
      return "已完成";
    case "failed":
      return "失败";
    case "queued":
      return "排队中";
    default:
      return status;
  }
}

function formatBytes(value: number): string {
  if (value <= 0) {
    return "0 KB";
  }
  if (value < 1024 * 1024) {
    return `${Math.ceil(value / 1024)} KB`;
  }
  return `${(value / 1024 / 1024).toFixed(1)} MB`;
}
