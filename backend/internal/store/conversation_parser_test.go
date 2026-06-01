package store

import (
	"testing"
	"time"
)

func TestParseUploadedConversationSplitsSenderTimeLines(t *testing.T) {
	messages := parseUploadedConversation("upl_test", "赵先生 10:02 你们这个方案适合我们20人的销售团队吗？\n销售A 10:04 适合，我先按您团队规模整理方案。", time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC))

	if len(messages) != 2 {
		t.Fatalf("expected two messages, got %d", len(messages))
	}
	if messages[0].SenderName != "赵先生" {
		t.Fatalf("expected customer sender, got %s", messages[0].SenderName)
	}
	if messages[0].SenderType != "customer" {
		t.Fatalf("expected customer sender type, got %s", messages[0].SenderType)
	}
	if messages[1].SenderName != "销售A" {
		t.Fatalf("expected sales sender, got %s", messages[1].SenderName)
	}
	if messages[1].SenderType != "sales" {
		t.Fatalf("expected sales sender type, got %s", messages[1].SenderType)
	}
}

func TestInferUploadedCustomerNameUsesFirstCustomerSender(t *testing.T) {
	name := inferUploadedCustomerName("销售A 10:00 您好\n李女士 10:02 我想了解价格")
	if name != "李女士" {
		t.Fatalf("expected first customer sender, got %s", name)
	}
}
