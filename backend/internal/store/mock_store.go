package store

import (
	"strings"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
)

type MockStore struct {
	customers []domain.Customer
	tasks     []domain.FollowupTask
}

func NewMockStore() *MockStore {
	customers := []domain.Customer{
		{
			ID:             "cus_001",
			Name:           "王女士",
			Source:         "企业微信",
			Owner:          domain.Owner{ID: "usr_001", Name: "销售A"},
			Stage:          domain.StagePriceObjection,
			Intent:         domain.IntentHigh,
			Concerns:       []string{"价格", "售后"},
			Tags:           []string{"价格敏感", "需要案例"},
			ProfileSummary: "关注价格和售后保障，近期购买意向较强。",
			LastContactAt:  "2026-05-28T09:30:00Z",
			PendingTasks:   1,
			RiskFlags:      []string{"涉及价格承诺，需人工确认"},
		},
		{
			ID:             "cus_002",
			Name:           "李先生",
			Source:         "上传聊天记录",
			Owner:          domain.Owner{ID: "usr_001", Name: "销售A"},
			Stage:          domain.StageSilent,
			Intent:         domain.IntentMedium,
			Concerns:       []string{"效果", "案例"},
			Tags:           []string{"需要案例"},
			ProfileSummary: "已了解产品但 3 天未回复，适合用案例轻触达。",
			LastContactAt:  "2026-05-25T18:20:00Z",
			PendingTasks:   1,
			RiskFlags:      []string{"客户沉默超过 72 小时"},
		},
		{
			ID:             "cus_003",
			Name:           "陈总",
			Source:         "企业微信",
			Owner:          domain.Owner{ID: "usr_002", Name: "销售B"},
			Stage:          domain.StageHighIntent,
			Intent:         domain.IntentHigh,
			Concerns:       []string{"交付周期", "售后"},
			Tags:           []string{"时间紧急", "关注售后"},
			ProfileSummary: "明确提出本周内确认方案，主管可关注推进。",
			LastContactAt:  "2026-05-28T08:50:00Z",
			PendingTasks:   1,
			RiskFlags:      []string{},
		},
	}

	return &MockStore{
		customers: customers,
		tasks: []domain.FollowupTask{
			newTask("task_001", customers[0], "price_objection", "2026-05-28T10:00:00Z", domain.AgentRecommendation{
				CustomerStage:     domain.StagePriceObjection,
				IntentLevel:       domain.IntentHigh,
				MainConcerns:      []string{"价格", "效果", "售后"},
				RecommendedAction: "解释方案价值并引导预约",
				Script:            "您好，刚才您提到比较关注价格，我这边帮您整理了一下更适合您的方案，也可以结合售后保障一起看。",
				Reasoning:         "客户连续询问价格和售后，说明有购买兴趣但存在决策顾虑。",
				RiskFlags:         []string{"涉及价格承诺，建议人工确认"},
				NextFollowupTime:  "2026-05-28T16:00:00Z",
			}),
			newTask("task_002", customers[1], "silent_reactivation", "2026-05-28T10:10:00Z", domain.AgentRecommendation{
				CustomerStage:     domain.StageSilent,
				IntentLevel:       domain.IntentMedium,
				MainConcerns:      []string{"效果", "案例"},
				RecommendedAction: "用案例轻触达",
				Script:            "您好，前面您提到比较关注实际效果，我整理了一个和您情况接近的案例，您方便的话我发您看一下。",
				Reasoning:         "客户已表达兴趣但长时间未回复，直接促单风险较高，适合先提供案例降低压力。",
				RiskFlags:         []string{"避免连续催促造成反感"},
				NextFollowupTime:  "2026-05-28T15:00:00Z",
			}),
			newTask("task_003", customers[2], "closing", "2026-05-28T10:20:00Z", domain.AgentRecommendation{
				CustomerStage:     domain.StageHighIntent,
				IntentLevel:       domain.IntentHigh,
				MainConcerns:      []string{"交付周期", "售后"},
				RecommendedAction: "确认决策节点并推动方案评审",
				Script:            "陈总，您这边如果希望本周确认，我建议今天先把交付周期和售后边界对齐，我可以整理成一页方案给您内部评估。",
				Reasoning:         "客户给出明确时间窗口，需要推进到具体评审动作。",
				RiskFlags:         []string{"交付周期不能超出实际能力承诺"},
				NextFollowupTime:  "2026-05-28T14:00:00Z",
			}),
		},
	}
}

func (s *MockStore) Customers() []domain.Customer {
	return append([]domain.Customer(nil), s.customers...)
}

func (s *MockStore) FollowupTasks() []domain.FollowupTask {
	return append([]domain.FollowupTask(nil), s.tasks...)
}

func newTask(id string, customer domain.Customer, taskType string, generatedAt string, recommendation domain.AgentRecommendation) domain.FollowupTask {
	return domain.FollowupTask{
		ID:             id,
		Customer:       customer,
		Type:           taskType,
		Status:         domain.FollowupPending,
		GeneratedAt:    generatedAt,
		Recommendation: recommendation,
		Feedback:       nil,
	}
}

func MatchCustomer(customer domain.Customer, keyword string, stage string, intent string, ownerID string, risk string) bool {
	if keyword != "" && !strings.Contains(customer.Name, keyword) {
		return false
	}
	if stage != "" && string(customer.Stage) != stage {
		return false
	}
	if intent != "" && string(customer.Intent) != intent {
		return false
	}
	if ownerID != "" && customer.Owner.ID != ownerID {
		return false
	}
	if risk == "1" && len(customer.RiskFlags) == 0 {
		return false
	}
	return true
}
