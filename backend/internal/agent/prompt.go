package agent

const (
	ModelMockLocalV1 = "mock-local-v1"

	TaskGenerateFollowupScript = "generate_followup_script"
	TaskRegenerateFollowup     = "regenerate_followup_script"

	PromptFollowupV1   = "followup_v1"
	PromptRegenerateV1 = "followup_regenerate_v1"
)

type PromptTemplate struct {
	Version        string
	TaskType       string
	SystemPrompt   string
	UserPromptHint string
	OutputSchema   string
	Description    string
}

func Template(version string) (PromptTemplate, bool) {
	template, ok := templates[version]
	return template, ok
}

var templates = map[string]PromptTemplate{
	PromptFollowupV1: {
		Version:        PromptFollowupV1,
		TaskType:       TaskGenerateFollowupScript,
		SystemPrompt:   "You are Qiling Agent, an enterprise private-domain sales assistant. Return only valid JSON matching the schema.",
		UserPromptHint: "Analyze the uploaded conversation, infer customer stage and intent, explain reasoning, list risks, and draft one follow-up script for human confirmation.",
		OutputSchema:   recommendationSchema,
		Description:    "Generate customer profile, intent judgment, risk flags, and a follow-up script from an uploaded conversation.",
	},
	PromptRegenerateV1: {
		Version:        PromptRegenerateV1,
		TaskType:       TaskRegenerateFollowup,
		SystemPrompt:   "You are Qiling Agent, an enterprise private-domain sales assistant. Return only valid JSON matching the schema.",
		UserPromptHint: "Rewrite the follow-up script according to user feedback while preserving customer context, reasoning, and risk controls.",
		OutputSchema:   recommendationSchema,
		Description:    "Regenerate the follow-up script from an existing task while preserving customer context and risk controls.",
	},
}

const recommendationSchema = `{
  "type": "object",
  "required": ["customer_stage", "intent_level", "main_concerns", "recommended_action", "script", "reasoning", "risk_flags"],
  "properties": {
    "customer_stage": {"type": "string"},
    "intent_level": {"type": "string"},
    "main_concerns": {"type": "array", "items": {"type": "string"}},
    "recommended_action": {"type": "string"},
    "script": {"type": "string"},
    "reasoning": {"type": "string"},
    "risk_flags": {"type": "array", "items": {"type": "string"}},
    "next_followup_time": {"type": "string"}
  }
}`
