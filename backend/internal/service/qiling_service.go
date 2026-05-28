package service

import (
	"net/http"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
	"github.com/heqiucheng/qiling-agent/backend/internal/store"
)

type QilingService struct {
	store *store.MockStore
}

func NewQilingService(store *store.MockStore) *QilingService {
	return &QilingService{store: store}
}

func (s *QilingService) DashboardSummary() domain.DashboardSummary {
	customers := s.store.Customers()
	tasks := s.store.FollowupTasks()

	return domain.DashboardSummary{
		Metrics: []domain.Metric{
			{Key: "pending_tasks", Label: "待确认话术", Value: len(tasks), Hint: "优先处理高意向客户"},
			{Key: "high_intent_customers", Label: "高意向客户", Value: countCustomersByIntent(customers, domain.IntentHigh), Hint: "建议今天完成跟进"},
			{Key: "risk_customers", Label: "风险客户", Value: countRiskCustomers(customers), Hint: "需要关注沉默和承诺风险"},
		},
		PriorityTasks:       tasks,
		HighIntentCustomers: filterCustomers(customers, func(c domain.Customer) bool { return c.Intent == domain.IntentHigh }),
		SilentCustomers:     filterCustomers(customers, func(c domain.Customer) bool { return c.Stage == domain.StageSilent }),
		RiskCustomers:       filterCustomers(customers, func(c domain.Customer) bool { return len(c.RiskFlags) > 0 }),
		DailyReview: domain.DailyReview{
			Summary:     "今天优先处理高意向客户和沉默客户，避免价格异议停留过久。",
			Suggestions: []string{"先跟进 2 小时内有互动的客户", "价格异议客户先给价值解释和案例，不直接承诺优惠"},
		},
	}
}

func (s *QilingService) Customers(r *http.Request) PageResult[domain.Customer] {
	query := r.URL.Query()
	customers := s.store.Customers()
	filtered := make([]domain.Customer, 0, len(customers))
	for _, customer := range customers {
		if store.MatchCustomer(
			customer,
			query.Get("keyword"),
			query.Get("stage"),
			query.Get("intent"),
			query.Get("owner_id"),
			query.Get("risk"),
		) {
			filtered = append(filtered, customer)
		}
	}
	return NewPageResult(filtered, PageRequestFromQuery(r))
}

func (s *QilingService) FollowupTasks(r *http.Request) PageResult[domain.FollowupTask] {
	query := r.URL.Query()
	tasks := s.store.FollowupTasks()
	filtered := make([]domain.FollowupTask, 0, len(tasks))
	for _, task := range tasks {
		if query.Get("status") != "" && string(task.Status) != query.Get("status") {
			continue
		}
		if query.Get("intent") != "" && string(task.Customer.Intent) != query.Get("intent") {
			continue
		}
		if query.Get("owner_id") != "" && task.Customer.Owner.ID != query.Get("owner_id") {
			continue
		}
		filtered = append(filtered, task)
	}
	return NewPageResult(filtered, PageRequestFromQuery(r))
}

func countCustomersByIntent(customers []domain.Customer, intent domain.IntentLevel) int {
	count := 0
	for _, customer := range customers {
		if customer.Intent == intent {
			count++
		}
	}
	return count
}

func countRiskCustomers(customers []domain.Customer) int {
	count := 0
	for _, customer := range customers {
		if len(customer.RiskFlags) > 0 {
			count++
		}
	}
	return count
}

func filterCustomers(customers []domain.Customer, keep func(domain.Customer) bool) []domain.Customer {
	filtered := make([]domain.Customer, 0, len(customers))
	for _, customer := range customers {
		if keep(customer) {
			filtered = append(filtered, customer)
		}
	}
	return filtered
}
