
package data

import "task/api/internal/entities"

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