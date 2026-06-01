package agent

const (
	ModelMockLocalV1 = "mock-local-v1"

	TaskGenerateFollowupScript = "generate_followup_script"
	TaskRegenerateFollowup     = "regenerate_followup_script"

	PromptFollowupV1   = "followup_v1"
	PromptRegenerateV1 = "followup_regenerate_v1"
)

type PromptTemplate struct {
	Version     string
	TaskType    string
	Description string
}

func Template(version string) (PromptTemplate, bool) {
	template, ok := templates[version]
	return template, ok
}

var templates = map[string]PromptTemplate{
	PromptFollowupV1: {
		Version:     PromptFollowupV1,
		TaskType:    TaskGenerateFollowupScript,
		Description: "Generate customer profile, intent judgment, risk flags, and a follow-up script from an uploaded conversation.",
	},
	PromptRegenerateV1: {
		Version:     PromptRegenerateV1,
		TaskType:    TaskRegenerateFollowup,
		Description: "Regenerate the follow-up script from an existing task while preserving customer context and risk controls.",
	},
}
