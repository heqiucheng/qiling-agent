package service

import (
	"bytes"
	"fmt"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
	"github.com/xuri/excelize/v2"
)

const reportXLSXContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

func renderReportXLSX(report domain.Report) ([]byte, error) {
	file := excelize.NewFile()
	defer func() {
		_ = file.Close()
	}()

	const summarySheet = "概览指标"
	const actionsSheet = "行动清单"
	const customersSheet = "客户明细"

	if err := file.SetSheetName("Sheet1", summarySheet); err != nil {
		return nil, err
	}
	if _, err := file.NewSheet(actionsSheet); err != nil {
		return nil, err
	}
	if _, err := file.NewSheet(customersSheet); err != nil {
		return nil, err
	}

	styles, err := newReportWorkbookStyles(file)
	if err != nil {
		return nil, err
	}
	if err := writeReportSummarySheet(file, styles, summarySheet, report); err != nil {
		return nil, err
	}
	if err := writeReportActionsSheet(file, styles, actionsSheet, report); err != nil {
		return nil, err
	}
	if err := writeReportCustomersSheet(file, styles, customersSheet, report); err != nil {
		return nil, err
	}
	file.SetActiveSheet(0)

	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

type reportWorkbookStyles struct {
	title  int
	header int
	body   int
	note   int
}

func newReportWorkbookStyles(file *excelize.File) (reportWorkbookStyles, error) {
	title, err := file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "111827"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return reportWorkbookStyles{}, err
	}
	header, err := file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"2563EB"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    reportThinBorders("D1D5DB"),
	})
	if err != nil {
		return reportWorkbookStyles{}, err
	}
	body, err := file.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
		Border:    reportThinBorders("E5E7EB"),
	})
	if err != nil {
		return reportWorkbookStyles{}, err
	}
	note, err := file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "4B5563"},
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
	})
	if err != nil {
		return reportWorkbookStyles{}, err
	}
	return reportWorkbookStyles{title: title, header: header, body: body, note: note}, nil
}

func reportThinBorders(color string) []excelize.Border {
	return []excelize.Border{
		{Type: "left", Color: color, Style: 1},
		{Type: "right", Color: color, Style: 1},
		{Type: "top", Color: color, Style: 1},
		{Type: "bottom", Color: color, Style: 1},
	}
}

func writeReportSummarySheet(file *excelize.File, styles reportWorkbookStyles, sheet string, report domain.Report) error {
	if err := file.MergeCell(sheet, "A1", "D1"); err != nil {
		return err
	}
	if err := file.SetCellValue(sheet, "A1", report.Title); err != nil {
		return err
	}
	if err := file.SetCellStyle(sheet, "A1", "A1", styles.title); err != nil {
		return err
	}
	if err := file.SetCellValue(sheet, "A2", report.Summary); err != nil {
		return err
	}
	if err := file.MergeCell(sheet, "A2", "D3"); err != nil {
		return err
	}
	if err := file.SetCellStyle(sheet, "A2", "D3", styles.note); err != nil {
		return err
	}

	rows := [][]any{{"指标", "数值", "中文备注", "生成时间"}}
	for _, metric := range report.Metrics {
		rows = append(rows, []any{metric.Label, metric.Value, metric.Hint, report.GeneratedAt})
	}
	if len(report.Metrics) == 0 {
		rows = append(rows, []any{"暂无指标", 0, "当前报告没有可汇总指标", report.GeneratedAt})
	}
	if err := writeRows(file, sheet, 5, rows); err != nil {
		return err
	}
	return formatTable(file, styles, sheet, 5, len(rows), []float64{20, 14, 42, 24})
}

func writeReportActionsSheet(file *excelize.File, styles reportWorkbookStyles, sheet string, report domain.Report) error {
	rows := [][]any{{"优先级", "客户", "建议动作", "时间提示"}}
	for _, item := range report.ActionItems {
		rows = append(rows, []any{item.Priority, item.CustomerName, item.Action, item.DueHint})
	}
	if len(report.ActionItems) == 0 {
		rows = append(rows, []any{"-", "-", "当前没有需要处理的行动项", "-"})
	}
	if err := writeRows(file, sheet, 1, rows); err != nil {
		return err
	}
	return formatTable(file, styles, sheet, 1, len(rows), []float64{14, 18, 60, 24})
}

func writeReportCustomersSheet(file *excelize.File, styles reportWorkbookStyles, sheet string, report domain.Report) error {
	rows := [][]any{{"板块", "客户", "阶段", "意愿", "建议动作", "推荐话术", "判断依据", "证据"}}
	for _, section := range report.Sections {
		for _, item := range section.Items {
			rows = append(rows, []any{
				section.Title,
				item.CustomerName,
				item.Stage,
				item.Intent,
				item.RecommendedAction,
				item.Script,
				item.Reasoning,
				joinReportEvidence(item.Evidence),
			})
		}
	}
	if len(rows) == 1 {
		rows = append(rows, []any{"-", "-", "-", "-", "当前没有客户明细", "-", "-", "-"})
	}
	if err := writeRows(file, sheet, 1, rows); err != nil {
		return err
	}
	return formatTable(file, styles, sheet, 1, len(rows), []float64{18, 18, 14, 14, 42, 52, 52, 64})
}

func writeRows(file *excelize.File, sheet string, startRow int, rows [][]any) error {
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, startRow+rowIndex)
			if err != nil {
				return err
			}
			if err := file.SetCellValue(sheet, cell, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func formatTable(file *excelize.File, styles reportWorkbookStyles, sheet string, startRow int, rowCount int, widths []float64) error {
	if rowCount == 0 {
		return nil
	}
	lastColName, err := excelize.ColumnNumberToName(len(widths))
	if err != nil {
		return err
	}
	headerStart := fmt.Sprintf("A%d", startRow)
	headerEnd := fmt.Sprintf("%s%d", lastColName, startRow)
	bodyStart := fmt.Sprintf("A%d", startRow+1)
	bodyEnd := fmt.Sprintf("%s%d", lastColName, startRow+rowCount-1)
	if err := file.SetCellStyle(sheet, headerStart, headerEnd, styles.header); err != nil {
		return err
	}
	if rowCount > 1 {
		if err := file.SetCellStyle(sheet, bodyStart, bodyEnd, styles.body); err != nil {
			return err
		}
	}
	for index, width := range widths {
		colName, err := excelize.ColumnNumberToName(index + 1)
		if err != nil {
			return err
		}
		if err := file.SetColWidth(sheet, colName, colName, width); err != nil {
			return err
		}
	}
	return file.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      startRow,
		TopLeftCell: fmt.Sprintf("A%d", startRow+1),
		ActivePane:  "bottomLeft",
	})
}

func joinReportEvidence(items []string) string {
	result := ""
	for index, item := range items {
		if index > 0 {
			result += "\n"
		}
		result += item
	}
	return result
}
