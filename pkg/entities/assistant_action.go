package entities

type RequestAssistantReaction struct {
	ActionType string `validate:"required"`
}

type AllowAssistantAction struct {
}

type AssistantAction struct {
	Action     int
	HasComment bool
	Comment    string
}

