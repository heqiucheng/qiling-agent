package store

import (
	"database/sql"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Customers() []domain.Customer {
	return []domain.Customer{}
}

func (r *MySQLRepository) Customer(id string) (domain.Customer, bool) {
	return domain.Customer{}, false
}

func (r *MySQLRepository) FollowupTasks() []domain.FollowupTask {
	return []domain.FollowupTask{}
}

func (r *MySQLRepository) FollowupTasksByCustomer(customerID string) []domain.FollowupTask {
	return []domain.FollowupTask{}
}

func (r *MySQLRepository) ConversationMessages(customerID string) []domain.ConversationMessage {
	return []domain.ConversationMessage{}
}

func (r *MySQLRepository) CreateUpload(sourceType string, content string, ownerID string) domain.UploadRecord {
	return domain.UploadRecord{}
}

func (r *MySQLRepository) Upload(id string) (domain.UploadRecord, bool) {
	return domain.UploadRecord{}, false
}

func (r *MySQLRepository) ConfirmUpload(uploadID string, customerName string, ownerID string) (domain.ConfirmUploadResult, error) {
	return domain.ConfirmUploadResult{}, nil
}

func (r *MySQLRepository) CopyTask(taskID string, copiedAt string) (domain.TaskCopyResult, error) {
	return domain.TaskCopyResult{}, nil
}

func (r *MySQLRepository) SkipTask(taskID string, reason string) (domain.TaskStatusResult, error) {
	return domain.TaskStatusResult{}, nil
}

func (r *MySQLRepository) MarkTaskWrong(taskID string, reason string) (domain.MarkWrongResult, error) {
	return domain.MarkWrongResult{}, nil
}

func (r *MySQLRepository) RegenerateTask(taskID string, instruction string) (domain.RegenerateTaskResult, error) {
	return domain.RegenerateTaskResult{}, nil
}
