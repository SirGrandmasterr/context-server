package presenter

type AssistantErrorResponse struct {
	Error string
}

func NewAssistantErrorResponse(err error) AssistantErrorResponse {
	return AssistantErrorResponse{
		Error: err.Error(),
	}
}
