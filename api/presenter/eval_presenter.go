package presenter

type TestErrorResponse struct {
	Error string
}

func NewTestErrorResponse(err error) ActionErrorResponse {
	return ActionErrorResponse{
		Error: err.Error(),
	}
}
