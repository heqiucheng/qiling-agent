package store

import "github.com/heqiucheng/qiling-agent/backend/internal/domain"

type Repository interface {
	Customers() []domain.Customer
	Customer(id string) (domain.Customer, bool)
	FollowupTasks() []domain.FollowupTask
	FollowupTasksByCustomer(customerID string) []domain.FollowupTask
	ConversationMessages(customerID string) []domain.ConversationMessage
	CreateUpload(sourceType string, content string, ownerID string) (domain.UploadRecord, error)
	Upload(id string) (domain.UploadRecord, bool)
	ConfirmUpload(uploadID string, customerName string, ownerID string) (domain.ConfirmUploadResult, error)
	CopyTask(taskID string, copiedAt string) (domain.TaskCopyResult, error)
	SkipTask(taskID string, reason string) (domain.TaskStatusResult, error)
	MarkTaskWrong(taskID string, reason string) (domain.MarkWrongResult, error)
	RegenerateTask(taskID string, instruction string) (domain.RegenerateTaskResult, error)
}
