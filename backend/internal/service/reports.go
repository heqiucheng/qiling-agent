package service

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/heqiucheng/qiling-agent/backend/internal/apperror"
	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
	"github.com/heqiucheng/qiling-agent/backend/internal/store"
)

type CustomerIntentReportRequest struct {
	Range string `json:"range"`
}

type ReportExport struct {
	Filename    string
	ContentType string
	Body        []byte
}

func (s *QilingService) CustomerIntentReport(req CustomerIntentReportRequest, actor domain.Actor) (domain.Report, error) {
	customers := visibleCustomers(s.store.Customers(), actor)
	tasks := visibleTasks(s.store.FollowupTasks(), actor)
	tasksByCustomer := latestTaskByCustomer(tasks)
	rangeLabel := reportRangeLabel(req.Range)
	now := time.Now().UTC()

	highIntentItems := make([]domain.ReportCustomerItem, 0)
	pendingItems := make([]domain.ReportCustomerItem, 0)
	riskItems := make([]domain.ReportCustomerItem, 0)
	clarifyItems := make([]domain.ReportCustomerItem, 0)
	actionItems := make([]domain.ReportActionItem, 0)

	for _, customer := range customers {
		task, hasTask := tasksByCustomer[customer.ID]
		item := reportCustomerItem(customer, task, hasTask, s.reportContextEvidence(customer.ID))
		switch {
		case isPendingReportCustomer(customer, item):
			pendingItems = append(pendingItems, item)
		case customer.Intent == domain.IntentHigh:
			highIntentItems = append(highIntentItems, item)
		case len(customer.RiskFlags) > 0 || customer.Stage == domain.StageSilent || customer.Stage == domain.StageChurnRisk:
			riskItems = append(riskItems, item)
		default:
			clarifyItems = append(clarifyItems, item)
		}
		actionItems = append(actionItems, reportActionItem(customer, item))
	}

	sections := []domain.ReportSection{
		reportSection("高意向客户", "已经出现明确推进信号，适合优先跟进。", highIntentItems),
		reportSection("待确认客户", "客户没有拒绝，但需要等待确认或补充信息。", pendingItems),
		reportSection("风险客户", "存在沉默、异议、承诺风险或体验风险，需要谨慎处理。", riskItems),
		reportSection("需要补充信息", "当前证据不足，先问清楚事项再推进。", clarifyItems),
	}
	sections = nonEmptyReportSections(sections)

	report := domain.Report{
		ID:         "rpt_customer_intent_" + now.Format("20060102150405"),
		Type:       domain.ReportCustomerIntent,
		Title:      rangeLabel + "客户意愿分析报告",
		RangeLabel: rangeLabel,
		OwnerID:    actor.UserID,
		OwnerRole:  actor.Role,
		Summary: fmt.Sprintf(
			"本报告共分析 %d 位可见客户，其中高意向 %d 位，待确认 %d 位，风险 %d 位，需要补充信息 %d 位。",
			len(customers),
			len(highIntentItems),
			len(pendingItems),
			len(riskItems),
			len(clarifyItems),
		),
		Metrics: []domain.Metric{
			{Key: "customers_total", Label: "分析客户", Value: len(customers), Hint: "当前账号可见客户数"},
			{Key: "high_intent", Label: "高意向", Value: len(highIntentItems), Hint: "建议优先推进"},
			{Key: "pending_confirmation", Label: "待确认", Value: len(pendingItems), Hint: "轻提醒，不要连续催"},
			{Key: "risk_customers", Label: "风险客户", Value: len(riskItems), Hint: "需要人工复核"},
		},
		Sections:    sections,
		ActionItems: actionItems,
		GeneratedAt: now.Format(time.RFC3339),
	}
	report.Markdown = renderReportMarkdown(report)
	return s.store.SaveReport(report)
}

func (s *QilingService) Reports(r *http.Request, actor domain.Actor) PageResult[domain.ReportSummary] {
	page := PageRequestFromQuery(r)
	result := s.store.ReportPage(actor.UserID, actor.Role, store.PageRequest{Page: page.Page, PageSize: page.PageSize})
	return NewPageResultWithTotal(result.Items, page, result.Total)
}

func (s *QilingService) Report(reportID string, actor domain.Actor) (domain.Report, error) {
	report, ok := s.store.Report(reportID)
	if !ok {
		return domain.Report{}, apperror.New("NOT_FOUND", "报告不存在", map[string]any{"report_id": reportID})
	}
	if report.OwnerID != actor.UserID || report.OwnerRole != actor.Role {
		return domain.Report{}, apperror.New("FORBIDDEN", "无权查看该报告", map[string]any{"report_id": reportID})
	}
	return report, nil
}

func (s *QilingService) ExportReport(reportID string, format string, actor domain.Actor) (ReportExport, error) {
	report, err := s.Report(reportID, actor)
	if err != nil {
		return ReportExport{}, err
	}
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "markdown"
	}
	switch format {
	case "markdown":
		return ReportExport{
			Filename:    safeReportExportFilename(report.ID) + ".md",
			ContentType: "text/markdown; charset=utf-8",
			Body:        []byte(report.Markdown),
		}, nil
	case "xlsx":
		body, err := renderReportXLSX(report)
		if err != nil {
			return ReportExport{}, err
		}
		return ReportExport{
			Filename:    safeReportExportFilename(report.ID) + ".xlsx",
			ContentType: reportXLSXContentType,
			Body:        body,
		}, nil
	default:
		return ReportExport{}, apperror.New("UNSUPPORTED_EXPORT_FORMAT", "暂不支持该报告导出格式", map[string]any{"format": format})
	}
}

func safeReportExportFilename(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "report"
	}
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", `"`, "_", "<", "_", ">", "_", "|", "_")
	return replacer.Replace(value)
}

func latestTaskByCustomer(tasks []domain.FollowupTask) map[string]domain.FollowupTask {
	result := make(map[string]domain.FollowupTask)
	for _, task := range tasks {
		current, ok := result[task.Customer.ID]
		if !ok || task.GeneratedAt > current.GeneratedAt {
			result[task.Customer.ID] = task
		}
	}
	return result
}

func (s *QilingService) reportContextEvidence(customerID string) []string {
	evidence := make([]string, 0, 8)
	runs := s.store.AgentRunsByCustomer(customerID, store.PageRequest{Page: 1, PageSize: 2}).Items
	for _, run := range runs {
		if strings.TrimSpace(run.Output.Reasoning) != "" {
			evidence = append(evidence, "AgentRun："+run.Output.Reasoning)
		} else if strings.TrimSpace(run.InputSummary) != "" {
			evidence = append(evidence, "AgentRun 输入："+run.InputSummary)
		}
		if len(run.RiskFlags) > 0 {
			evidence = append(evidence, "AgentRun 风险："+strings.Join(run.RiskFlags, "、"))
		}
	}

	messages := s.store.ConversationMessagePage(customerID, store.PageRequest{Page: 1, PageSize: 3}).Items
	for _, message := range messages {
		if strings.TrimSpace(message.Content) != "" {
			evidence = append(evidence, "最近聊天："+message.SenderName+"："+message.Content)
		}
	}

	facts := s.store.LongTermMemoryFacts(customerID, store.PageRequest{Page: 1, PageSize: 3}).Items
	for _, fact := range facts {
		if strings.TrimSpace(fact.Value) != "" {
			evidence = append(evidence, "长期记忆："+fact.Category+"."+fact.Key+"="+fact.Value)
		}
	}
	if len(facts) == 0 {
		evidence = append(evidence, "长期记忆：暂无已确认长期事实")
	}
	return compactEvidence(evidence)
}

func reportCustomerItem(customer domain.Customer, task domain.FollowupTask, hasTask bool, contextEvidence []string) domain.ReportCustomerItem {
	action := "先补充确认客户当前需求，再安排下一步跟进。"
	script := "我先确认一下，您现在主要想核对哪一项？您把不确定的地方发我，我这边帮您一起看清楚。"
	reasoning := customer.ProfileSummary
	if strings.TrimSpace(reasoning) == "" {
		reasoning = "当前客户信息不足，需要补充沟通证据。"
	}
	evidence := reportEvidence(customer)
	if hasTask {
		action = firstNonEmptyReportValue(task.Recommendation.RecommendedAction, action)
		script = firstNonEmptyReportValue(task.Recommendation.Script, script)
		reasoning = firstNonEmptyReportValue(task.Recommendation.Reasoning, reasoning)
		evidence = append(evidence, "最新任务建议："+task.Recommendation.RecommendedAction)
	}
	evidence = append(evidence, contextEvidence...)
	return domain.ReportCustomerItem{
		CustomerID:        customer.ID,
		CustomerName:      customer.Name,
		Stage:             string(customer.Stage),
		Intent:            string(customer.Intent),
		RecommendedAction: action,
		Script:            script,
		Reasoning:         reasoning,
		Evidence:          compactEvidence(evidence),
	}
}

func reportEvidence(customer domain.Customer) []string {
	evidence := make([]string, 0, 4)
	if strings.TrimSpace(customer.ProfileSummary) != "" {
		evidence = append(evidence, "画像："+customer.ProfileSummary)
	}
	if len(customer.Concerns) > 0 {
		evidence = append(evidence, "关注点："+strings.Join(customer.Concerns, "、"))
	}
	if len(customer.RiskFlags) > 0 {
		evidence = append(evidence, "风险："+strings.Join(customer.RiskFlags, "、"))
	}
	if strings.TrimSpace(customer.LastContactAt) != "" {
		evidence = append(evidence, "最近联系："+customer.LastContactAt)
	}
	if len(evidence) == 0 {
		evidence = append(evidence, "信息不足，需要补充确认")
	}
	return evidence
}

func compactEvidence(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
		if len(result) >= 8 {
			break
		}
	}
	return result
}

func isPendingReportCustomer(customer domain.Customer, item domain.ReportCustomerItem) bool {
	text := strings.ToLower(strings.Join(append([]string{customer.ProfileSummary, item.Reasoning, item.RecommendedAction}, customer.Concerns...), " "))
	signals := []string{"pending", "confirm", "确认", "待确认", "明天", "稍后", "晚点", "核对"}
	for _, signal := range signals {
		if strings.Contains(text, signal) {
			return true
		}
	}
	return customer.Stage == domain.StageNeedsDiscovery && customer.Intent == domain.IntentMedium
}

func reportActionItem(customer domain.Customer, item domain.ReportCustomerItem) domain.ReportActionItem {
	priority := "medium"
	dueHint := "今天内确认下一步"
	if customer.Intent == domain.IntentHigh {
		priority = "high"
		dueHint = "优先跟进"
	}
	if len(customer.RiskFlags) > 0 || customer.Stage == domain.StageSilent || customer.Stage == domain.StageChurnRisk {
		priority = "risk"
		dueHint = "先复核风险再触达"
	}
	if isPendingReportCustomer(customer, item) {
		priority = "pending"
		dueHint = "按对方承诺时间轻提醒"
	}
	return domain.ReportActionItem{
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		Priority:     priority,
		Action:       item.RecommendedAction,
		DueHint:      dueHint,
	}
}

func reportSection(title string, summary string, items []domain.ReportCustomerItem) domain.ReportSection {
	evidence := make([]string, 0)
	if len(items) == 0 {
		evidence = append(evidence, "当前没有符合该分组的客户。")
	} else {
		for _, item := range items {
			if len(item.Evidence) > 0 {
				evidence = append(evidence, item.CustomerName+"："+item.Evidence[0])
			}
			if len(evidence) >= 3 {
				break
			}
		}
	}
	return domain.ReportSection{Title: title, Summary: summary, Items: items, Evidence: evidence}
}

func nonEmptyReportSections(sections []domain.ReportSection) []domain.ReportSection {
	result := make([]domain.ReportSection, 0, len(sections))
	for _, section := range sections {
		if len(section.Items) > 0 {
			result = append(result, section)
		}
	}
	if len(result) == 0 {
		return sections[:1]
	}
	return result
}

func reportRangeLabel(value string) string {
	switch strings.TrimSpace(value) {
	case "today":
		return "今日"
	case "last_30_days":
		return "最近 30 天"
	case "last_7_days", "":
		return "最近 7 天"
	default:
		return strings.TrimSpace(value)
	}
}

func firstNonEmptyReportValue(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func renderReportMarkdown(report domain.Report) string {
	var builder strings.Builder
	builder.WriteString("# " + report.Title + "\n\n")
	builder.WriteString(report.Summary + "\n\n")
	builder.WriteString("## 指标\n\n")
	for _, metric := range report.Metrics {
		builder.WriteString(fmt.Sprintf("- %s：%v（%s）\n", metric.Label, metric.Value, metric.Hint))
	}
	builder.WriteString("\n")
	for _, section := range report.Sections {
		builder.WriteString("## " + section.Title + "\n\n")
		builder.WriteString(section.Summary + "\n\n")
		for _, item := range section.Items {
			builder.WriteString(fmt.Sprintf("### %s\n\n", item.CustomerName))
			builder.WriteString("- 状态：" + item.Stage + "\n")
			builder.WriteString("- 意向：" + item.Intent + "\n")
			builder.WriteString("- 建议动作：" + item.RecommendedAction + "\n")
			builder.WriteString("- 推荐话术：" + item.Script + "\n")
			builder.WriteString("- 判断依据：" + item.Reasoning + "\n")
			if len(item.Evidence) > 0 {
				builder.WriteString("- 证据：" + strings.Join(item.Evidence, "；") + "\n")
			}
			builder.WriteString("\n")
		}
	}
	builder.WriteString("## 行动清单\n\n")
	if len(report.ActionItems) == 0 {
		builder.WriteString("- 当前没有需要处理的行动项。\n")
	} else {
		for _, item := range report.ActionItems {
			builder.WriteString(fmt.Sprintf("- [%s] %s：%s（%s）\n", item.Priority, item.CustomerName, item.Action, item.DueHint))
		}
	}
	return strings.TrimSpace(builder.String())
}
