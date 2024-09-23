package entities

type RequestAssistantReaction struct {
	ActionName string `validate:"required"`
}

type AssistantAction struct {
	Action     int
	HasComment bool
	Comment    string
}
