package service

import (
	"strings"
	"testing"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
)

func TestRenderReportPDF(t *testing.T) {
	body, err := renderReportPDF(domain.Report{
		ID:          "rpt_test",
		Title:       "测试报告",
		RangeLabel:  "最近 7 天",
		Summary:     "测试摘要",
		GeneratedAt: "2026-06-02T00:00:00Z",
		Metrics: []domain.Metric{
			{Label: "客户数", Value: 1, Hint: "测试"},
		},
		Sections: []domain.ReportSection{
			{
				Title:   "高意向客户",
				Summary: "测试分组",
				Items: []domain.ReportCustomerItem{
					{
						CustomerName:      "李总",
						Stage:             "qualified",
						Intent:            "high",
						RecommendedAction: "跟进",
						Script:            "您好",
						Reasoning:         "有明确意向",
						Evidence:          []string{"聊天记录"},
					},
				},
			},
		},
		ActionItems: []domain.ReportActionItem{
			{Priority: "high", CustomerName: "李总", Action: "跟进", DueHint: "今天"},
		},
	})
	if err != nil {
		t.Fatalf("render pdf: %v", err)
	}
	if !strings.HasPrefix(string(body), "%PDF") {
		t.Fatalf("expected pdf header")
	}
}
