package store

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/heqiucheng/qiling-agent/backend/internal/agent"
	"github.com/heqiucheng/qiling-agent/backend/internal/db"
	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
)

func integrationRepository(t *testing.T) *MySQLRepository {
	t.Helper()

	dsn := os.Getenv("QILING_INTEGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("set QILING_INTEGRATION_DATABASE_URL to run MySQL integration tests")
	}

	database, err := db.OpenMySQL(dsn)
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})

	migrationsDir := filepath.Join("..", "..", "migrations")
	if _, err := db.Reset(database, migrationsDir); err != nil {
		t.Fatalf("reset mysql: %v", err)
	}

	return NewMySQLRepository(database)
}

func testConfirmAgentRun() ConfirmUploadAgentRun {
	recommendation := domain.AgentRecommendation{
		CustomerStage:     domain.StagePriceObjection,
		IntentLevel:       domain.IntentHigh,
		MainConcerns:      []string{"price", "effect"},
		RecommendedAction: "explain value and provide a similar case",
		Script:            "The customer mentioned price and effect, so explain value before pushing the next step.",
		Reasoning:         "The uploaded content shows price and effect concerns.",
		RiskFlags:         []string{"avoid direct discount or outcome promises"},
	}
	return ConfirmUploadAgentRun{
		TaskType:         agent.TaskGenerateFollowupScript,
		Model:            agent.ModelMockLocalV1,
		PromptVersion:    agent.PromptFollowupV1,
		InputSummary:     "test uploaded conversation",
		Recommendation:   recommendation,
		ValidationErrors: agent.ValidateRecommendation(recommendation),
		RiskFlags:        recommendation.RiskFlags,
	}
}

func TestMySQLRepositoryConfirmUploadIsIdempotent(t *testing.T) {
	repository := integrationRepository(t)

	upload, err := repository.CreateUpload("pasted_text", "Customer A 10:20 price and effect need review", "usr_001")
	if err != nil {
		t.Fatalf("create upload: %v", err)
	}

	first, err := repository.ConfirmUpload(upload.ID, "Customer A", "usr_001", testConfirmAgentRun())
	if err != nil {
		t.Fatalf("first confirm: %v", err)
	}
	second, err := repository.ConfirmUpload(upload.ID, "Customer A", "usr_001", testConfirmAgentRun())
	if err != nil {
		t.Fatalf("second confirm should be idempotent: %v", err)
	}

	if first.CustomerID != second.CustomerID {
		t.Fatalf("expected same customer id, got %s and %s", first.CustomerID, second.CustomerID)
	}
	if first.FollowupTaskID != second.FollowupTaskID {
		t.Fatalf("expected same task id, got %s and %s", first.FollowupTaskID, second.FollowupTaskID)
	}
	if first.AgentRunID != second.AgentRunID {
		t.Fatalf("expected same agent run id, got %s and %s", first.AgentRunID, second.AgentRunID)
	}
}

func TestMySQLRepositoryTaskStatusUpdateAllowsOneConcurrentWinner(t *testing.T) {
	repository := integrationRepository(t)

	upload, err := repository.CreateUpload("pasted_text", "Customer B 10:20 price and effect need review", "usr_001")
	if err != nil {
		t.Fatalf("create upload: %v", err)
	}
	confirm, err := repository.ConfirmUpload(upload.ID, "Customer B", "usr_001", testConfirmAgentRun())
	if err != nil {
		t.Fatalf("confirm upload: %v", err)
	}

	const workers = 8
	var wg sync.WaitGroup
	results := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repository.CopyTask(confirm.FollowupTaskID, "2026-06-01T10:00:00Z")
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	successes := 0
	failures := 0
	for err := range results {
		if err == nil {
			successes++
			continue
		}
		failures++
	}

	if successes != 1 {
		t.Fatalf("expected one successful status transition, got %d successes and %d failures", successes, failures)
	}
	if failures != workers-1 {
		t.Fatalf("expected remaining requests to fail, got %d", failures)
	}
}

func TestMySQLRepositoryPagesListQueries(t *testing.T) {
	repository := integrationRepository(t)

	customers := repository.CustomerPage(CustomerFilter{Intent: string(domain.IntentHigh)}, PageRequest{Page: 1, PageSize: 1})
	if customers.Total != 2 {
		t.Fatalf("expected two high intent customers, got %d", customers.Total)
	}
	if len(customers.Items) != 1 {
		t.Fatalf("expected one customer page item, got %d", len(customers.Items))
	}

	tasks := repository.FollowupTaskPage(FollowupTaskFilter{Status: string(domain.FollowupPending)}, PageRequest{Page: 1, PageSize: 2})
	if tasks.Total != 3 {
		t.Fatalf("expected three pending tasks, got %d", tasks.Total)
	}
	if len(tasks.Items) != 2 {
		t.Fatalf("expected two task page items, got %d", len(tasks.Items))
	}

	messages := repository.ConversationMessagePage("cus_001", PageRequest{Page: 1, PageSize: 1})
	if messages.Total != 2 {
		t.Fatalf("expected two messages, got %d", messages.Total)
	}
	if len(messages.Items) != 1 {
		t.Fatalf("expected one message page item, got %d", len(messages.Items))
	}
}

func TestMySQLRepositoryPersistsAuditEvents(t *testing.T) {
	repository := integrationRepository(t)

	created, err := repository.CreateAuditEvent(domain.AuditEvent{
		Action:     domain.AuditFollowupTaskCopied,
		Actor:      domain.Actor{UserID: "usr_001", Role: "sales"},
		RequestID:  "req_test",
		EntityType: "followup_task",
		EntityID:   "task_001",
		Metadata: map[string]any{
			"has_script": true,
		},
	})
	if err != nil {
		t.Fatalf("create audit event: %v", err)
	}

	page := repository.AuditEventPage(AuditEventFilter{
		Action:   string(domain.AuditFollowupTaskCopied),
		ActorID:  "usr_001",
		EntityID: "task_001",
	}, PageRequest{Page: 1, PageSize: 10})

	if page.Total != 1 {
		t.Fatalf("expected one audit event, got %d", page.Total)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected one audit item, got %d", len(page.Items))
	}
	if page.Items[0].ID != created.ID {
		t.Fatalf("expected audit id %s, got %s", created.ID, page.Items[0].ID)
	}
	if page.Items[0].Metadata["has_script"] != true {
		t.Fatalf("expected metadata has_script true, got %#v", page.Items[0].Metadata["has_script"])
	}
}

func TestMySQLRepositoryAgentRunLookup(t *testing.T) {
	repository := integrationRepository(t)

	upload, err := repository.CreateUpload("pasted_text", "Customer A 10:20 price and effect need review", "usr_001")
	if err != nil {
		t.Fatalf("create upload: %v", err)
	}
	confirm, err := repository.ConfirmUpload(upload.ID, "Customer A", "usr_001", testConfirmAgentRun())
	if err != nil {
		t.Fatalf("confirm upload: %v", err)
	}

	run, ok := repository.AgentRun(confirm.AgentRunID)
	if !ok {
		t.Fatalf("expected agent run %s", confirm.AgentRunID)
	}
	if run.Model != "mock-local-v1" {
		t.Fatalf("expected mock model, got %s", run.Model)
	}
	if run.PromptVersion != "followup_v1" {
		t.Fatalf("expected followup_v1, got %s", run.PromptVersion)
	}
	if run.Output.Script == "" {
		t.Fatal("expected structured output script")
	}
}

func TestMySQLRepositoryAgentRunsByCustomer(t *testing.T) {
	repository := integrationRepository(t)

	upload, err := repository.CreateUpload("pasted_text", "Customer A 10:20 price and effect need review", "usr_001")
	if err != nil {
		t.Fatalf("create upload: %v", err)
	}
	confirm, err := repository.ConfirmUpload(upload.ID, "Customer A", "usr_001", testConfirmAgentRun())
	if err != nil {
		t.Fatalf("confirm upload: %v", err)
	}

	page := repository.AgentRunsByCustomer(confirm.CustomerID, PageRequest{Page: 1, PageSize: 5})
	if page.Total != 1 {
		t.Fatalf("expected one agent run, got %d", page.Total)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected one page item, got %d", len(page.Items))
	}
	if page.Items[0].ID != confirm.AgentRunID {
		t.Fatalf("expected agent run %s, got %s", confirm.AgentRunID, page.Items[0].ID)
	}
}

func TestMySQLRepositoryUpsertsLongTermMemoryFacts(t *testing.T) {
	repository := integrationRepository(t)

	upload, err := repository.CreateUpload("pasted_text", "Customer A 10:20 price and effect need review", "usr_001")
	if err != nil {
		t.Fatalf("create upload: %v", err)
	}
	confirm, err := repository.ConfirmUpload(upload.ID, "Customer A", "usr_001", testConfirmAgentRun())
	if err != nil {
		t.Fatalf("confirm upload: %v", err)
	}

	first, err := repository.UpsertLongTermMemoryFact(domain.LongTermMemoryFact{
		CustomerID: confirm.CustomerID,
		Category:   "profile",
		Key:        "intent_level",
		Value:      "high",
		Confidence: 0.8,
		SourceType: "agent_run",
		SourceID:   confirm.AgentRunID,
		Status:     domain.MemoryFactActive,
	})
	if err != nil {
		t.Fatalf("upsert first memory fact: %v", err)
	}
	second, err := repository.UpsertLongTermMemoryFact(domain.LongTermMemoryFact{
		CustomerID: confirm.CustomerID,
		Category:   "profile",
		Key:        "intent_level",
		Value:      "medium",
		Confidence: 0.6,
		SourceType: "agent_run",
		SourceID:   confirm.AgentRunID,
		Status:     domain.MemoryFactActive,
	})
	if err != nil {
		t.Fatalf("upsert second memory fact: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("expected upsert to preserve fact id %s, got %s", first.ID, second.ID)
	}

	page := repository.LongTermMemoryFacts(confirm.CustomerID, PageRequest{Page: 1, PageSize: 10})
	if page.Total != 1 {
		t.Fatalf("expected one active fact, got %d", page.Total)
	}
	if page.Items[0].Value != "medium" {
		t.Fatalf("expected updated value medium, got %s", page.Items[0].Value)
	}
}

func TestMySQLRepositoryRejectsLongTermMemoryFact(t *testing.T) {
	repository := integrationRepository(t)

	upload, err := repository.CreateUpload("pasted_text", "Customer A 10:20 price and effect need review", "usr_001")
	if err != nil {
		t.Fatalf("create upload: %v", err)
	}
	confirm, err := repository.ConfirmUpload(upload.ID, "Customer A", "usr_001", testConfirmAgentRun())
	if err != nil {
		t.Fatalf("confirm upload: %v", err)
	}
	fact, err := repository.UpsertLongTermMemoryFact(domain.LongTermMemoryFact{
		CustomerID: confirm.CustomerID,
		Category:   "concern",
		Key:        "price",
		Value:      "price",
		Confidence: 0.8,
		SourceType: "agent_run",
		SourceID:   confirm.AgentRunID,
		Status:     domain.MemoryFactActive,
	})
	if err != nil {
		t.Fatalf("upsert memory fact: %v", err)
	}

	result, err := repository.UpdateLongTermMemoryFactStatus(confirm.CustomerID, fact.ID, domain.MemoryFactRejected)
	if err != nil {
		t.Fatalf("reject memory fact: %v", err)
	}
	if result.Status != domain.MemoryFactRejected {
		t.Fatalf("expected rejected status, got %s", result.Status)
	}

	page := repository.LongTermMemoryFacts(confirm.CustomerID, PageRequest{Page: 1, PageSize: 10})
	if page.Total != 0 {
		t.Fatalf("expected rejected fact to be excluded from active facts, got %d", page.Total)
	}
}

func TestMySQLRepositoryCorrectsLongTermMemoryFact(t *testing.T) {
	repository := integrationRepository(t)

	upload, err := repository.CreateUpload("pasted_text", "Customer A 10:20 price and effect need review", "usr_001")
	if err != nil {
		t.Fatalf("create upload: %v", err)
	}
	confirm, err := repository.ConfirmUpload(upload.ID, "Customer A", "usr_001", testConfirmAgentRun())
	if err != nil {
		t.Fatalf("confirm upload: %v", err)
	}
	fact, err := repository.UpsertLongTermMemoryFact(domain.LongTermMemoryFact{
		CustomerID: confirm.CustomerID,
		Category:   "concern",
		Key:        "price",
		Value:      "price",
		Confidence: 0.8,
		SourceType: "agent_run",
		SourceID:   confirm.AgentRunID,
		Status:     domain.MemoryFactActive,
	})
	if err != nil {
		t.Fatalf("upsert memory fact: %v", err)
	}

	result, err := repository.CorrectLongTermMemoryFact(confirm.CustomerID, fact.ID, domain.LongTermMemoryFact{
		CustomerID: confirm.CustomerID,
		Category:   "concern",
		Key:        "delivery",
		Value:      "delivery timeline",
		Confidence: 1,
		SourceType: "human_correction",
		SourceID:   fact.ID,
		Status:     domain.MemoryFactActive,
	})
	if err != nil {
		t.Fatalf("correct memory fact: %v", err)
	}
	if result.OldStatus != domain.MemoryFactSuperseded {
		t.Fatalf("expected old fact superseded, got %s", result.OldStatus)
	}
	if result.NewFact.Value != "delivery timeline" {
		t.Fatalf("expected corrected value, got %s", result.NewFact.Value)
	}

	page := repository.LongTermMemoryFacts(confirm.CustomerID, PageRequest{Page: 1, PageSize: 10})
	if page.Total != 1 {
		t.Fatalf("expected one active corrected fact, got %d", page.Total)
	}
	if page.Items[0].Key != "delivery" {
		t.Fatalf("expected active corrected key delivery, got %s", page.Items[0].Key)
	}
}
