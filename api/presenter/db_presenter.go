package presenter

type ActionErrorResponse struct {
	Error string
}

func NewActionErrorResponse(err error) ActionErrorResponse {
	return ActionErrorResponse{
		Error: err.Error(),
	}
}
