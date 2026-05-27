package task

type TaskResponse struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
    Description string `json:"description"`
    Status      string `json:"status"`
	AssigneeID  *int    `json:"assignee_id"`
}


