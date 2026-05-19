package response

type ApiResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func SuccessResponse(message string, data interface{}) ApiResponse {
	return ApiResponse{
		Status:  true,
		Message: message,
		Data:    data,
	}
}

func ErrorResponse(message string, err any) ApiResponse {
	return ApiResponse{
		Status:  false,
		Message: message,
		Error:   err,
	}
}
