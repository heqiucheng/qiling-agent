package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/heqiucheng/qiling-agent/backend/internal/apperror"
	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Customers() []domain.Customer {
	rows, err := r.db.Query(`
		SELECT c.id, c.name, c.source, c.owner_id, u.name, c.stage, c.intent,
		       c.concerns, c.tags, c.profile_summary, c.last_contact_at,
		       c.pending_tasks, c.risk_flags
		FROM customers c
		INNER JOIN users u ON u.id = c.owner_id
		ORDER BY c.last_contact_at DESC, c.id DESC
	`)
	if err != nil {
		return []domain.Customer{}
	}
	defer rows.Close()

	return scanCustomers(rows)
}

func (r *MySQLRepository) Customer(id string) (domain.Customer, bool) {
	row := r.db.QueryRow(`
		SELECT c.id, c.name, c.source, c.owner_id, u.name, c.stage, c.intent,
		       c.concerns, c.tags, c.profile_summary, c.last_contact_at,
		       c.pending_tasks, c.risk_flags
		FROM customers c
		INNER JOIN users u ON u.id = c.owner_id
		WHERE c.id = ?
	`, id)

	customer, err := scanCustomer(row)
	if err != nil {
		return domain.Customer{}, false
	}
	return customer, true
}

func (r *MySQLRepository) FollowupTasks() []domain.FollowupTask {
	rows, err := r.db.Query(followupTaskSelectSQL() + " ORDER BY t.generated_at DESC, t.id DESC")
	if err != nil {
		return []domain.FollowupTask{}
	}
	defer rows.Close()

	return scanFollowupTasks(rows)
}

func (r *MySQLRepository) FollowupTasksByCustomer(customerID string) []domain.FollowupTask {
	rows, err := r.db.Query(followupTaskSelectSQL()+" WHERE t.customer_id = ? ORDER BY t.generated_at DESC, t.id DESC", customerID)
	if err != nil {
		return []domain.FollowupTask{}
	}
	defer rows.Close()

	return scanFollowupTasks(rows)
}

func (r *MySQLRepository) ConversationMessages(customerID string) []domain.ConversationMessage {
	rows, err := r.db.Query(`
		SELECT id, sender_type, sender_name, content, sent_at
		FROM conversation_messages
		WHERE customer_id = ?
		ORDER BY sent_at ASC, id ASC
	`, customerID)
	if err != nil {
		return []domain.ConversationMessage{}
	}
	defer rows.Close()

	messages := make([]domain.ConversationMessage, 0)
	for rows.Next() {
		var message domain.ConversationMessage
		var sentAt time.Time
		if err := rows.Scan(&message.ID, &message.SenderType, &message.SenderName, &message.Content, &sentAt); err != nil {
			return []domain.ConversationMessage{}
		}
		message.SentAt = formatTime(sentAt)
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return []domain.ConversationMessage{}
	}
	return messages
}

func (r *MySQLRepository) CreateUpload(sourceType string, content string, ownerID string) (domain.UploadRecord, error) {
	return domain.UploadRecord{}, unsupportedMySQLWrite("CreateUpload")
}

func (r *MySQLRepository) Upload(id string) (domain.UploadRecord, bool) {
	return domain.UploadRecord{}, false
}

func (r *MySQLRepository) ConfirmUpload(uploadID string, customerName string, ownerID string) (domain.ConfirmUploadResult, error) {
	return domain.ConfirmUploadResult{}, unsupportedMySQLWrite("ConfirmUpload")
}

func (r *MySQLRepository) CopyTask(taskID string, copiedAt string) (domain.TaskCopyResult, error) {
	return domain.TaskCopyResult{}, unsupportedMySQLWrite("CopyTask")
}

func (r *MySQLRepository) SkipTask(taskID string, reason string) (domain.TaskStatusResult, error) {
	return domain.TaskStatusResult{}, unsupportedMySQLWrite("SkipTask")
}

func (r *MySQLRepository) MarkTaskWrong(taskID string, reason string) (domain.MarkWrongResult, error) {
	return domain.MarkWrongResult{}, unsupportedMySQLWrite("MarkTaskWrong")
}

func (r *MySQLRepository) RegenerateTask(taskID string, instruction string) (domain.RegenerateTaskResult, error) {
	return domain.RegenerateTaskResult{}, unsupportedMySQLWrite("RegenerateTask")
}

type customerScanner interface {
	Scan(dest ...any) error
}

func scanCustomers(rows *sql.Rows) []domain.Customer {
	customers := make([]domain.Customer, 0)
	for rows.Next() {
		customer, err := scanCustomer(rows)
		if err != nil {
			return []domain.Customer{}
		}
		customers = append(customers, customer)
	}
	if err := rows.Err(); err != nil {
		return []domain.Customer{}
	}
	return customers
}

func scanCustomer(scanner customerScanner) (domain.Customer, error) {
	var customer domain.Customer
	var owner domain.Owner
	var stage string
	var intent string
	var concernsJSON []byte
	var tagsJSON []byte
	var riskFlagsJSON []byte
	var lastContactAt time.Time

	if err := scanner.Scan(
		&customer.ID,
		&customer.Name,
		&customer.Source,
		&owner.ID,
		&owner.Name,
		&stage,
		&intent,
		&concernsJSON,
		&tagsJSON,
		&customer.ProfileSummary,
		&lastContactAt,
		&customer.PendingTasks,
		&riskFlagsJSON,
	); err != nil {
		return domain.Customer{}, err
	}

	customer.Owner = owner
	customer.Stage = domain.CustomerStage(stage)
	customer.Intent = domain.IntentLevel(intent)
	customer.LastContactAt = formatTime(lastContactAt)
	customer.Concerns = decodeStringList(concernsJSON)
	customer.Tags = decodeStringList(tagsJSON)
	customer.RiskFlags = decodeStringList(riskFlagsJSON)
	return customer, nil
}

func followupTaskSelectSQL() string {
	return `
		SELECT t.id, t.type, t.status, t.generated_at, t.recommendation, t.feedback,
		       c.id, c.name, c.source, c.owner_id, u.name, c.stage, c.intent,
		       c.concerns, c.tags, c.profile_summary, c.last_contact_at,
		       c.pending_tasks, c.risk_flags
		FROM followup_tasks t
		INNER JOIN customers c ON c.id = t.customer_id
		INNER JOIN users u ON u.id = c.owner_id
	`
}

func scanFollowupTasks(rows *sql.Rows) []domain.FollowupTask {
	tasks := make([]domain.FollowupTask, 0)
	for rows.Next() {
		var task domain.FollowupTask
		var status string
		var generatedAt time.Time
		var recommendationJSON []byte
		var feedbackJSON sql.NullString
		var customer domain.Customer
		var owner domain.Owner
		var stage string
		var intent string
		var concernsJSON []byte
		var tagsJSON []byte
		var riskFlagsJSON []byte
		var lastContactAt time.Time

		if err := rows.Scan(
			&task.ID,
			&task.Type,
			&status,
			&generatedAt,
			&recommendationJSON,
			&feedbackJSON,
			&customer.ID,
			&customer.Name,
			&customer.Source,
			&owner.ID,
			&owner.Name,
			&stage,
			&intent,
			&concernsJSON,
			&tagsJSON,
			&customer.ProfileSummary,
			&lastContactAt,
			&customer.PendingTasks,
			&riskFlagsJSON,
		); err != nil {
			return []domain.FollowupTask{}
		}

		customer.Owner = owner
		customer.Stage = domain.CustomerStage(stage)
		customer.Intent = domain.IntentLevel(intent)
		customer.Concerns = decodeStringList(concernsJSON)
		customer.Tags = decodeStringList(tagsJSON)
		customer.LastContactAt = formatTime(lastContactAt)
		customer.RiskFlags = decodeStringList(riskFlagsJSON)

		task.Customer = customer
		task.Status = domain.FollowupTaskStatus(status)
		task.GeneratedAt = formatTime(generatedAt)
		task.Recommendation = decodeRecommendation(recommendationJSON)
		if feedbackJSON.Valid {
			feedback := domain.TaskFeedback{}
			if err := json.Unmarshal([]byte(feedbackJSON.String), &feedback); err == nil {
				task.Feedback = &feedback
			}
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return []domain.FollowupTask{}
	}
	return tasks
}

func decodeStringList(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}

	values := make([]string, 0)
	if err := json.Unmarshal(raw, &values); err != nil {
		return []string{}
	}
	return values
}

func decodeRecommendation(raw []byte) domain.AgentRecommendation {
	var recommendation domain.AgentRecommendation
	if len(raw) == 0 {
		return recommendation
	}
	if err := json.Unmarshal(raw, &recommendation); err != nil {
		return domain.AgentRecommendation{}
	}
	return recommendation
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func unsupportedMySQLWrite(method string) error {
	return apperror.New("MYSQL_WRITE_NOT_READY", "MySQL 写入能力尚未启用", map[string]any{"method": method})
}
