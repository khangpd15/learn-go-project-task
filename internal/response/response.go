
package response

type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
func SuccessResponse(
	message string,
	data interface{},
) ApiResponse {

	return ApiResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}
func ErrorResponse(message string) ApiResponse {

	return ApiResponse{
		Success: false,
		Message: message,
	}
}