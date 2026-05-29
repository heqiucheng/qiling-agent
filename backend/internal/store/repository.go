package store

import "github.com/heqiucheng/qiling-agent/backend/internal/domain"

type PageRequest struct {
	Page     int
	PageSize int
}

type CustomerFilter struct {
	Keyword string
	Stage   string
	Intent  string
	OwnerID string
	Risk    string
}

type CustomerPage struct {
	Items []domain.Customer
	Total int
}

type FollowupTaskFilter struct {
	Status  string
	Intent  string
	OwnerID string
}

type FollowupTaskPage struct {
	Items []domain.FollowupTask
	Total int
}

type ConversationMessagePage struct {
	Items []domain.ConversationMessage
	Total int
}

type Repository interface {
	Customers() []domain.Customer
	CustomerPage(filter CustomerFilter, page PageRequest) CustomerPage
	Customer(id string) (domain.Customer, bool)
	FollowupTasks() []domain.FollowupTask
	FollowupTaskPage(filter FollowupTaskFilter, page PageRequest) FollowupTaskPage
	FollowupTasksByCustomer(customerID string) []domain.FollowupTask
	ConversationMessages(customerID string) []domain.ConversationMessage
	ConversationMessagePage(customerID string, page PageRequest) ConversationMessagePage
	CreateUpload(sourceType string, content string, ownerID string) (domain.UploadRecord, error)
	Upload(id string) (domain.UploadRecord, bool)
	ConfirmUpload(uploadID string, customerName string, ownerID string) (domain.ConfirmUploadResult, error)
	CopyTask(taskID string, copiedAt string) (domain.TaskCopyResult, error)
	SkipTask(taskID string, reason string) (domain.TaskStatusResult, error)
	MarkTaskWrong(taskID string, reason string) (domain.MarkWrongResult, error)
	RegenerateTask(taskID string, instruction string) (domain.RegenerateTaskResult, error)
}
