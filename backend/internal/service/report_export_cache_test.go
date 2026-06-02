package service

import (
	"testing"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
	"github.com/heqiucheng/qiling-agent/backend/internal/store"
)

func TestExportReportUsesDefensiveCacheCopies(t *testing.T) {
	service := NewQilingService(store.NewMockStore())
	actor := domain.Actor{UserID: "usr_001", Role: "sales"}
	report, err := service.CustomerIntentReport(CustomerIntentReportRequest{Range: "last_7_days"}, actor)
	if err != nil {
		t.Fatalf("create report: %v", err)
	}

	first, err := service.ExportReport(report.ID, "pdf", actor)
	if err != nil {
		t.Fatalf("first export: %v", err)
	}
	if len(first.Body) == 0 {
		t.Fatalf("expected export body")
	}
	first.Body[0] = 'X'

	second, err := service.ExportReport(report.ID, "pdf", actor)
	if err != nil {
		t.Fatalf("second export: %v", err)
	}
	if len(second.Body) == 0 || second.Body[0] == 'X' {
		t.Fatalf("expected cached export to be protected from caller mutation")
	}
}
