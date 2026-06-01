package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
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

func (r *MySQLRepository) CustomerPage(filter CustomerFilter, page PageRequest) CustomerPage {
	where, args := customerWhere(filter)
	total := countRows(r.db, "SELECT COUNT(*) FROM customers c INNER JOIN users u ON u.id = c.owner_id"+where, args)

	query := `
		SELECT c.id, c.name, c.source, c.owner_id, u.name, c.stage, c.intent,
		       c.concerns, c.tags, c.profile_summary, c.last_contact_at,
		       c.pending_tasks, c.risk_flags
		FROM customers c
		INNER JOIN users u ON u.id = c.owner_id
	` + where + " ORDER BY c.last_contact_at DESC, c.id DESC LIMIT ? OFFSET ?"
	args = append(args, page.PageSize, pageOffset(page))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return CustomerPage{Items: []domain.Customer{}, Total: 0}
	}
	defer rows.Close()

	return CustomerPage{Items: scanCustomers(rows), Total: total}
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

func (r *MySQLRepository) FollowupTaskPage(filter FollowupTaskFilter, page PageRequest) FollowupTaskPage {
	where, args := followupTaskWhere(filter)
	total := countRows(r.db, "SELECT COUNT(*) FROM followup_tasks t INNER JOIN customers c ON c.id = t.customer_id INNER JOIN users u ON u.id = c.owner_id"+where, args)

	query := followupTaskSelectSQL() + where + " ORDER BY t.generated_at DESC, t.id DESC LIMIT ? OFFSET ?"
	args = append(args, page.PageSize, pageOffset(page))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return FollowupTaskPage{Items: []domain.FollowupTask{}, Total: 0}
	}
	defer rows.Close()

	return FollowupTaskPage{Items: scanFollowupTasks(rows), Total: total}
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

func (r *MySQLRepository) ConversationMessagePage(customerID string, page PageRequest) ConversationMessagePage {
	total := countRows(r.db, "SELECT COUNT(*) FROM conversation_messages WHERE customer_id = ?", []any{customerID})

	rows, err := r.db.Query(`
		SELECT id, sender_type, sender_name, content, sent_at
		FROM conversation_messages
		WHERE customer_id = ?
		ORDER BY sent_at ASC, id ASC
		LIMIT ? OFFSET ?
	`, customerID, page.PageSize, pageOffset(page))
	if err != nil {
		return ConversationMessagePage{Items: []domain.ConversationMessage{}, Total: 0}
	}
	defer rows.Close()

	return ConversationMessagePage{Items: scanConversationMessages(rows), Total: total}
}

func (r *MySQLRepository) CreateUpload(sourceType string, content string, ownerID string) (domain.UploadRecord, error) {
	now := time.Now().UTC()
	id := makeID("upl", now)
	customerName := inferCustomerName(content)
	record := domain.UploadRecord{
		ID:         id,
		Status:     domain.UploadNeedsConfirmation,
		SourceType: sourceType,
		ParsedCustomer: domain.ParsedCustomer{
			Name:      customerName,
			OwnerName: ownerName(ownerID),
		},
		Messages: []domain.ConversationMessage{
			{
				ID:         "msg_" + id,
				SenderType: "customer",
				SenderName: customerName,
				Content:    strings.TrimSpace(content),
				SentAt:     formatTime(now),
			},
		},
		Warnings:  []string{},
		CreatedAt: formatTime(now),
	}

	warningsJSON, err := json.Marshal(record.Warnings)
	if err != nil {
		return domain.UploadRecord{}, err
	}

	_, err = r.db.Exec(`
		INSERT INTO uploads (id, status, source_type, raw_content, parsed_customer_name, parsed_owner_name, warnings, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, record.ID, record.Status, record.SourceType, strings.TrimSpace(content), record.ParsedCustomer.Name, record.ParsedCustomer.OwnerName, warningsJSON, now)
	if err != nil {
		return domain.UploadRecord{}, err
	}

	return record, nil
}

func (r *MySQLRepository) Upload(id string) (domain.UploadRecord, bool) {
	var record domain.UploadRecord
	var status string
	var warningsJSON []byte
	var createdAt time.Time
	var rawContent sql.NullString

	err := r.db.QueryRow(`
		SELECT id, status, source_type, parsed_customer_name, parsed_owner_name, warnings, created_at, raw_content
		FROM uploads
		WHERE id = ?
	`, id).Scan(
		&record.ID,
		&status,
		&record.SourceType,
		&record.ParsedCustomer.Name,
		&record.ParsedCustomer.OwnerName,
		&warningsJSON,
		&createdAt,
		&rawContent,
	)
	if err != nil {
		return domain.UploadRecord{}, false
	}

	record.Status = domain.UploadStatus(status)
	record.Warnings = decodeStringList(warningsJSON)
	record.CreatedAt = formatTime(createdAt)
	record.Messages = []domain.ConversationMessage{
		{
			ID:         "msg_" + record.ID,
			SenderType: "customer",
			SenderName: record.ParsedCustomer.Name,
			Content:    rawContent.String,
			SentAt:     record.CreatedAt,
		},
	}
	return record, true
}

func (r *MySQLRepository) ConfirmUpload(uploadID string, customerName string, ownerID string) (domain.ConfirmUploadResult, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return domain.ConfirmUploadResult{}, err
	}
	defer tx.Rollback()

	record, err := uploadForUpdate(tx, uploadID)
	if err != nil {
		return domain.ConfirmUploadResult{}, err
	}
	if record.Status == domain.UploadConfirmed {
		result, ok := confirmedUploadResult(record)
		if ok {
			return result, nil
		}
		return domain.ConfirmUploadResult{}, apperror.New("CONFLICT", "上传记录已确认", map[string]any{"upload_id": uploadID})
	}
	if strings.TrimSpace(customerName) == "" {
		customerName = record.ParsedCustomer.Name
	}

	now := time.Now().UTC()
	customerID := makeID("cus", now)
	taskID := makeID("task", now)
	agentRunID := makeID("run", now)
	conversationID := makeID("conv", now)

	concernsJSON := mustJSON([]string{"价格", "效果"})
	tagsJSON := mustJSON([]string{"价格敏感"})
	riskFlagsJSON := mustJSON([]string{"涉及价格承诺，需人工确认"})
	recommendation := domain.AgentRecommendation{
		CustomerStage:     domain.StagePriceObjection,
		IntentLevel:       domain.IntentHigh,
		MainConcerns:      []string{"价格", "效果"},
		RecommendedAction: "解释价值并提供案例",
		Script:            "您好，您刚才提到价格和效果，我建议先结合您的使用场景看投入产出，我可以给您整理一个接近情况的案例。",
		Reasoning:         "上传内容显示客户关注价格和效果，需要先建立价值感再推动下一步。",
		RiskFlags:         []string{"避免直接承诺优惠或效果"},
		NextFollowupTime:  formatTime(now.Add(6 * time.Hour)),
	}
	recommendationJSON := mustJSON(recommendation)

	if _, err := tx.Exec(`
		INSERT INTO customers (
			id, name, source, owner_id, stage, intent, concerns, tags,
			profile_summary, last_contact_at, pending_tasks, risk_flags
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, customerID, customerName, "上传聊天记录", ownerID, domain.StagePriceObjection, domain.IntentHigh, concernsJSON, tagsJSON, "由上传聊天记录生成的客户画像，客户正在比较价格和效果。", now, 1, riskFlagsJSON); err != nil {
		return domain.ConfirmUploadResult{}, err
	}

	if _, err := tx.Exec(`
		INSERT INTO conversation_messages (id, customer_id, sender_type, sender_name, content, sent_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "msg_"+uploadID, customerID, "customer", customerName, "上传聊天记录已解析，原文将在后续解析器中结构化保存。", now); err != nil {
		return domain.ConfirmUploadResult{}, err
	}

	if _, err := tx.Exec(`
		INSERT INTO followup_tasks (id, customer_id, type, status, generated_at, recommendation, feedback)
		VALUES (?, ?, ?, ?, ?, ?, NULL)
	`, taskID, customerID, "price_objection", domain.FollowupPending, now, recommendationJSON); err != nil {
		return domain.ConfirmUploadResult{}, err
	}

	if _, err := tx.Exec(`
		INSERT INTO agent_runs (
			id, customer_id, task_type, status, model, prompt_version, input_summary,
			output, validation_errors, risk_flags, created_at, completed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, agentRunID, customerID, "generate_followup_script", "succeeded", "mock-local-v1", "followup_v1", "上传聊天记录生成客户画像和跟进话术", recommendationJSON, mustJSON([]string{}), riskFlagsJSON, now, now); err != nil {
		return domain.ConfirmUploadResult{}, err
	}

	if _, err := tx.Exec(`
		UPDATE uploads
		SET status = ?,
		    confirmed_customer_id = ?,
		    conversation_id = ?,
		    agent_run_id = ?,
		    followup_task_id = ?
		WHERE id = ?
	`, domain.UploadConfirmed, customerID, conversationID, agentRunID, taskID, uploadID); err != nil {
		return domain.ConfirmUploadResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.ConfirmUploadResult{}, err
	}

	return domain.ConfirmUploadResult{
		CustomerID:     customerID,
		ConversationID: conversationID,
		AgentRunID:     agentRunID,
		FollowupTaskID: taskID,
		Status:         domain.UploadConfirmed,
	}, nil
}

func (r *MySQLRepository) CopyTask(taskID string, copiedAt string) (domain.TaskCopyResult, error) {
	if err := r.updatePendingTaskStatus(taskID, domain.FollowupCopied, nil); err != nil {
		return domain.TaskCopyResult{}, err
	}
	return domain.TaskCopyResult{TaskID: taskID, Status: domain.FollowupCopied, CopiedAt: copiedAt}, nil
}

func (r *MySQLRepository) SkipTask(taskID string, reason string) (domain.TaskStatusResult, error) {
	feedback := &domain.TaskFeedback{Reason: reason}
	if err := r.updatePendingTaskStatus(taskID, domain.FollowupSkipped, feedback); err != nil {
		return domain.TaskStatusResult{}, err
	}
	return domain.TaskStatusResult{TaskID: taskID, Status: domain.FollowupSkipped}, nil
}

func (r *MySQLRepository) MarkTaskWrong(taskID string, reason string) (domain.MarkWrongResult, error) {
	feedback := &domain.TaskFeedback{Reason: reason}
	if err := r.updatePendingTaskStatus(taskID, domain.FollowupMarkedWrong, feedback); err != nil {
		return domain.MarkWrongResult{}, err
	}
	return domain.MarkWrongResult{TaskID: taskID, Status: domain.FollowupMarkedWrong, FeedbackID: "fb_" + taskID}, nil
}

func (r *MySQLRepository) RegenerateTask(taskID string, instruction string) (domain.RegenerateTaskResult, error) {
	task, err := r.followupTask(taskID)
	if err != nil {
		return domain.RegenerateTaskResult{}, err
	}

	recommendation := task.Recommendation
	if strings.TrimSpace(instruction) != "" {
		recommendation.Script = recommendation.Script + "（已按反馈调整语气）"
		recommendation.Reasoning = recommendation.Reasoning + " 本次换一种话术保留原客户上下文，仅调整表达方式。"
	}

	recommendationJSON := mustJSON(recommendation)
	if _, err := r.db.Exec(`UPDATE followup_tasks SET recommendation = ? WHERE id = ?`, recommendationJSON, taskID); err != nil {
		return domain.RegenerateTaskResult{}, err
	}

	agentRunID := "run_" + taskID
	return domain.RegenerateTaskResult{TaskID: taskID, AgentRunID: agentRunID, Recommendation: recommendation}, nil
}

func (r *MySQLRepository) CreateAuditEvent(event domain.AuditEvent) (domain.AuditEvent, error) {
	now := time.Now().UTC()
	if event.ID == "" {
		event.ID = makeID("audit", now)
	}
	if event.CreatedAt == "" {
		event.CreatedAt = formatTime(now)
	}
	if event.Metadata == nil {
		event.Metadata = map[string]any{}
	}

	createdAt, err := time.Parse(time.RFC3339, event.CreatedAt)
	if err != nil {
		createdAt = now
		event.CreatedAt = formatTime(now)
	}
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return domain.AuditEvent{}, err
	}

	_, err = r.db.Exec(`
		INSERT INTO audit_events (
			id, action, actor_user_id, actor_role, request_id, entity_type, entity_id,
			related_type, related_id, metadata, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, event.ID, event.Action, event.Actor.UserID, event.Actor.Role, event.RequestID, event.EntityType, event.EntityID, nullIfEmpty(event.RelatedType), nullIfEmpty(event.RelatedID), metadataJSON, createdAt)
	if err != nil {
		return domain.AuditEvent{}, err
	}

	return event, nil
}

func (r *MySQLRepository) AuditEventPage(filter AuditEventFilter, page PageRequest) AuditEventPage {
	where, args := auditEventWhere(filter)
	total := countRows(r.db, "SELECT COUNT(*) FROM audit_events"+where, args)

	query := `
		SELECT id, action, actor_user_id, actor_role, request_id, entity_type, entity_id,
		       related_type, related_id, metadata, created_at
		FROM audit_events
	` + where + " ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?"
	args = append(args, page.PageSize, pageOffset(page))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return AuditEventPage{Items: []domain.AuditEvent{}, Total: 0}
	}
	defer rows.Close()

	return AuditEventPage{Items: scanAuditEvents(rows), Total: total}
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

func scanConversationMessages(rows *sql.Rows) []domain.ConversationMessage {
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

type mysqlUploadRecord struct {
	domain.UploadRecord
	ConfirmedCustomerID sql.NullString
	ConversationID      sql.NullString
	AgentRunID          sql.NullString
	FollowupTaskID      sql.NullString
}

func uploadForUpdate(tx *sql.Tx, id string) (mysqlUploadRecord, error) {
	var record mysqlUploadRecord
	var status string
	var warningsJSON []byte
	var createdAt time.Time
	var rawContent sql.NullString

	err := tx.QueryRow(`
		SELECT id, status, source_type, parsed_customer_name, parsed_owner_name, warnings, created_at, raw_content,
		       confirmed_customer_id, conversation_id, agent_run_id, followup_task_id
		FROM uploads
		WHERE id = ?
		FOR UPDATE
	`, id).Scan(
		&record.ID,
		&status,
		&record.SourceType,
		&record.ParsedCustomer.Name,
		&record.ParsedCustomer.OwnerName,
		&warningsJSON,
		&createdAt,
		&rawContent,
		&record.ConfirmedCustomerID,
		&record.ConversationID,
		&record.AgentRunID,
		&record.FollowupTaskID,
	)
	if err == sql.ErrNoRows {
		return mysqlUploadRecord{}, apperror.New("NOT_FOUND", "上传记录不存在", map[string]any{"upload_id": id})
	}
	if err != nil {
		return mysqlUploadRecord{}, err
	}

	record.Status = domain.UploadStatus(status)
	record.Warnings = decodeStringList(warningsJSON)
	record.CreatedAt = formatTime(createdAt)
	record.Messages = []domain.ConversationMessage{
		{
			ID:         "msg_" + record.ID,
			SenderType: "customer",
			SenderName: record.ParsedCustomer.Name,
			Content:    rawContent.String,
			SentAt:     record.CreatedAt,
		},
	}
	return record, nil
}

func confirmedUploadResult(record mysqlUploadRecord) (domain.ConfirmUploadResult, bool) {
	if !record.ConfirmedCustomerID.Valid || !record.ConversationID.Valid || !record.AgentRunID.Valid || !record.FollowupTaskID.Valid {
		return domain.ConfirmUploadResult{}, false
	}
	return domain.ConfirmUploadResult{
		CustomerID:     record.ConfirmedCustomerID.String,
		ConversationID: record.ConversationID.String,
		AgentRunID:     record.AgentRunID.String,
		FollowupTaskID: record.FollowupTaskID.String,
		Status:         domain.UploadConfirmed,
	}, true
}

func (r *MySQLRepository) followupTask(taskID string) (domain.FollowupTask, error) {
	rows, err := r.db.Query(followupTaskSelectSQL()+" WHERE t.id = ?", taskID)
	if err != nil {
		return domain.FollowupTask{}, err
	}
	defer rows.Close()

	tasks := scanFollowupTasks(rows)
	if len(tasks) == 0 {
		return domain.FollowupTask{}, apperror.New("NOT_FOUND", "跟进任务不存在", map[string]any{"task_id": taskID})
	}
	return tasks[0], nil
}

func (r *MySQLRepository) updatePendingTaskStatus(taskID string, status domain.FollowupTaskStatus, feedback *domain.TaskFeedback) error {
	var feedbackValue any
	if feedback != nil {
		feedbackValue = mustJSON(feedback)
	}

	result, err := r.db.Exec(`
		UPDATE followup_tasks
		SET status = ?, feedback = ?
		WHERE id = ? AND status = ?
	`, status, feedbackValue, taskID, domain.FollowupPending)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected > 0 {
		return nil
	}

	task, err := r.followupTask(taskID)
	if err != nil {
		return err
	}
	if task.Status != domain.FollowupPending {
		return apperror.New("TASK_ALREADY_FINALIZED", "任务已经处理，不能重复操作", map[string]any{"task_id": taskID})
	}

	return apperror.New("CONFLICT", "任务状态更新冲突，请刷新后重试", map[string]any{"task_id": taskID})
}

func makeID(prefix string, now time.Time) string {
	return fmt.Sprintf("%s_%d", prefix, now.UnixNano())
}

func mustJSON(value any) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		return []byte("null")
	}
	return raw
}

func customerWhere(filter CustomerFilter) (string, []any) {
	conditions := make([]string, 0)
	args := make([]any, 0)

	if filter.Keyword != "" {
		conditions = append(conditions, "c.name LIKE ?")
		args = append(args, "%"+filter.Keyword+"%")
	}
	if filter.Stage != "" {
		conditions = append(conditions, "c.stage = ?")
		args = append(args, filter.Stage)
	}
	if filter.Intent != "" {
		conditions = append(conditions, "c.intent = ?")
		args = append(args, filter.Intent)
	}
	if filter.OwnerID != "" {
		conditions = append(conditions, "c.owner_id = ?")
		args = append(args, filter.OwnerID)
	}
	if filter.Risk == "1" {
		conditions = append(conditions, "JSON_LENGTH(c.risk_flags) > 0")
	}

	return whereClause(conditions), args
}

func followupTaskWhere(filter FollowupTaskFilter) (string, []any) {
	conditions := make([]string, 0)
	args := make([]any, 0)

	if filter.Status != "" {
		conditions = append(conditions, "t.status = ?")
		args = append(args, filter.Status)
	}
	if filter.Intent != "" {
		conditions = append(conditions, "c.intent = ?")
		args = append(args, filter.Intent)
	}
	if filter.OwnerID != "" {
		conditions = append(conditions, "c.owner_id = ?")
		args = append(args, filter.OwnerID)
	}

	return whereClause(conditions), args
}

func whereClause(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(conditions, " AND ")
}

func pageOffset(page PageRequest) int {
	return (page.Page - 1) * page.PageSize
}

func countRows(db *sql.DB, query string, args []any) int {
	var total int
	if err := db.QueryRow(query, args...).Scan(&total); err != nil {
		return 0
	}
	return total
}

func scanAuditEvents(rows *sql.Rows) []domain.AuditEvent {
	events := make([]domain.AuditEvent, 0)
	for rows.Next() {
		var event domain.AuditEvent
		var action string
		var actor domain.Actor
		var relatedType sql.NullString
		var relatedID sql.NullString
		var metadataJSON []byte
		var createdAt time.Time

		if err := rows.Scan(
			&event.ID,
			&action,
			&actor.UserID,
			&actor.Role,
			&event.RequestID,
			&event.EntityType,
			&event.EntityID,
			&relatedType,
			&relatedID,
			&metadataJSON,
			&createdAt,
		); err != nil {
			return []domain.AuditEvent{}
		}

		event.Action = domain.AuditAction(action)
		event.Actor = actor
		event.RelatedType = relatedType.String
		event.RelatedID = relatedID.String
		event.Metadata = decodeObject(metadataJSON)
		event.CreatedAt = formatTime(createdAt)
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return []domain.AuditEvent{}
	}
	return events
}

func decodeObject(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	value := map[string]any{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return map[string]any{}
	}
	return value
}

func auditEventWhere(filter AuditEventFilter) (string, []any) {
	conditions := make([]string, 0)
	args := make([]any, 0)

	if filter.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, filter.Action)
	}
	if filter.ActorID != "" {
		conditions = append(conditions, "actor_user_id = ?")
		args = append(args, filter.ActorID)
	}
	if filter.EntityType != "" {
		conditions = append(conditions, "entity_type = ?")
		args = append(args, filter.EntityType)
	}
	if filter.EntityID != "" {
		conditions = append(conditions, "entity_id = ?")
		args = append(args, filter.EntityID)
	}

	return whereClause(conditions), args
}

func nullIfEmpty(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}
