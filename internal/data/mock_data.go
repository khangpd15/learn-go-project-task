
package data

import "task_api/internal/entities"

var Tasks = []entities.Task{
	{
		ID:          1,
		Title:       "Learn Go",
		Description: "Study Go basics",
		Status:      "TODO",
		Assignee:    "Khang",
	},
		{
		ID:          2,
		Title:       "Build a REST API",
		Description: "Create a simple REST API in Go",
		Status:      "IN_PROGRESS",
		Assignee:    "Han",
	},
		{
		ID:          3,
		Title:       "Write Documentation",
		Description: "Document the REST API",
		Status:      "DONE",
		Assignee:    "Duy",
	},
}
var User = []entities.User{
	{
		ID:    1,
		Username:  "Admin",
		Email: "admin@gmail.com",
		Role:  "ADMIN",
	},
	{
		ID:    2,
		Username:  "Khang",
		Email: "khang@gmail.com",
		Role:  "CUSTOMER",
	},
	{
		ID:    3,
		Username:  "Guest",
		Email: "guest@gmail.com",
		Role:  "GUEST",
	},
}