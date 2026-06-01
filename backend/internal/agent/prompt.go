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
		SystemPrompt:   "You are Qiling Agent, an enterprise conversation analyst. You handle sales, service, work coordination, life events, relationship maintenance, schedule confirmation, emotional context, and low-information chats. Return only one valid JSON object matching the schema. Every field must use the required type: reasoning must be a string, main_concerns and risk_flags must be string arrays.",
		UserPromptHint: "First identify the conversation scene before judging intent. Do not force every chat into sales, price objection, or deal-closing logic. If the chat is about daily life, salary, documents, work handoff, schedule confirmation, relationship maintenance, or emotional support, produce a follow-up that fits that scene. If evidence is limited, use medium or low intent, explain uncertainty, and recommend a light confirmation rather than aggressive selling.",
		OutputSchema:   recommendationSchema,
		Description:    "Generate scene-aware profile, intent judgment, risk flags, and a follow-up script from an uploaded conversation.",
	},
	PromptRegenerateV1: {
		Version:        PromptRegenerateV1,
		TaskType:       TaskRegenerateFollowup,
		SystemPrompt:   "You are Qiling Agent, an enterprise conversation analyst. Return only one valid JSON object matching the schema. Every field must use the required type: reasoning must be a string, main_concerns and risk_flags must be string arrays.",
		UserPromptHint: "Rewrite the follow-up script according to user feedback while preserving the real conversation scene, customer context, reasoning, and risk controls. Do not turn non-sales life or work coordination chats into sales scripts unless the evidence supports it.",
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
