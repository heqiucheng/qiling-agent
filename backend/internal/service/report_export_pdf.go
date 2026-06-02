package service

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
	"github.com/jung-kurt/gofpdf"
)

const reportPDFContentType = "application/pdf"

func renderReportPDF(report domain.Report) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(16, 16, 16)
	pdf.SetAutoPageBreak(true, 16)
	fontFamily := configureReportPDFFont(pdf)
	pdf.AddPage()

	pdf.SetFont(fontFamily, "B", 18)
	writePDFText(pdf, 0, 9, report.Title)
	pdf.SetFont(fontFamily, "", 10)
	writePDFText(pdf, 0, 6, report.RangeLabel+" · "+report.GeneratedAt)
	writePDFText(pdf, 0, 6, report.Summary)

	writePDFHeading(pdf, fontFamily, "指标概览")
	for _, metric := range report.Metrics {
		writePDFText(pdf, 0, 6, metric.Label+"："+toReportPDFText(metric.Value)+"。"+metric.Hint)
	}

	writePDFHeading(pdf, fontFamily, "客户分析")
	for _, section := range report.Sections {
		writePDFHeading(pdf, fontFamily, section.Title)
		writePDFText(pdf, 0, 6, section.Summary)
		for _, item := range section.Items {
			pdf.SetFont(fontFamily, "B", 11)
			writePDFText(pdf, 0, 6, item.CustomerName+" ｜ "+item.Intent+" ｜ "+item.Stage)
			pdf.SetFont(fontFamily, "", 10)
			writePDFText(pdf, 0, 6, "建议动作："+item.RecommendedAction)
			writePDFText(pdf, 0, 6, "推荐话术："+item.Script)
			writePDFText(pdf, 0, 6, "判断依据："+item.Reasoning)
			if len(item.Evidence) > 0 {
				writePDFText(pdf, 0, 6, "证据："+strings.Join(item.Evidence, "；"))
			}
			pdf.Ln(1.5)
		}
	}

	writePDFHeading(pdf, fontFamily, "行动清单")
	if len(report.ActionItems) == 0 {
		writePDFText(pdf, 0, 6, "当前没有需要处理的行动项。")
	} else {
		for _, item := range report.ActionItems {
			writePDFText(pdf, 0, 6, "["+item.Priority+"] "+item.CustomerName+"："+item.Action+"（"+item.DueHint+"）")
		}
	}

	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func configureReportPDFFont(pdf *gofpdf.Fpdf) string {
	candidates := []string{
		os.Getenv("QILING_PDF_FONT_PATH"),
		`C:\Windows\Fonts\simhei.ttf`,
		`C:\Windows\Fonts\msyh.ttf`,
		`C:\Windows\Fonts\msyh.ttc`,
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
	}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			pdf.AddUTF8Font("qiling", "", candidate)
			pdf.AddUTF8Font("qiling", "B", candidate)
			pdf.SetFont("qiling", "", 10)
			return "qiling"
		}
	}
	pdf.SetFont("Helvetica", "", 10)
	return "Helvetica"
}

func writePDFHeading(pdf *gofpdf.Fpdf, fontFamily string, text string) {
	pdf.Ln(3)
	pdf.SetFont(fontFamily, "B", 13)
	writePDFText(pdf, 0, 7, text)
	pdf.SetFont(fontFamily, "", 10)
}

func writePDFText(pdf *gofpdf.Fpdf, width float64, height float64, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	pdf.MultiCell(width, height, text, "", "L", false)
}

func toReportPDFText(value any) string {
	return fmt.Sprint(value)
}
