package store

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

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

func TestMySQLRepositoryConfirmUploadIsIdempotent(t *testing.T) {
	repository := integrationRepository(t)

	upload, err := repository.CreateUpload("pasted_text", "王女士 10:20 价格和效果需要再看看", "usr_001")
	if err != nil {
		t.Fatalf("create upload: %v", err)
	}

	first, err := repository.ConfirmUpload(upload.ID, "王女士", "usr_001")
	if err != nil {
		t.Fatalf("first confirm: %v", err)
	}
	second, err := repository.ConfirmUpload(upload.ID, "王女士", "usr_001")
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

	upload, err := repository.CreateUpload("pasted_text", "李先生 10:20 价格和效果需要再看看", "usr_001")
	if err != nil {
		t.Fatalf("create upload: %v", err)
	}
	confirm, err := repository.ConfirmUpload(upload.ID, "李先生", "usr_001")
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
