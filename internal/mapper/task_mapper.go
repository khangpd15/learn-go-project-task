package mapper

import (
	TaskRequest "task_api/internal/dto/request/task"
	TaskResponse "task_api/internal/dto/response/task"
	"task_api/internal/entities"
)

func ToTaskResponse(taskEntity entities.Task) TaskResponse.TaskResponse {
	return TaskResponse.TaskResponse{
		ID:          taskEntity.ID,
		Title:       taskEntity.Title,
		Description: taskEntity.Description,
		Status:      taskEntity.Status,
		AssigneeID:  taskEntity.AssigneeID,
	}
}
func TasksToResponses(taskEntities []entities.Task) []TaskResponse.TaskResponse {
	responses := make([]TaskResponse.TaskResponse, len(taskEntities))
	for i, task := range taskEntities {
		responses[i] = ToTaskResponse(task)
	}
	return responses
}
func CreateTaskRequestToTaskEntity(createReq TaskRequest.CreateTaskRequest) entities.Task {
	return entities.Task{
		ProjectID:   createReq.ProjectID,
		Title:       createReq.Title,
		Description: createReq.Description,
		Status:      "TODO", // Default status for new tasks
		AssigneeID: nil,
		
	}
}
func UpdateTaskRequestToTaskEntity(updateReq TaskRequest.UpdateTaskRequest) entities.Task {
	task := entities.Task{}

	if updateReq.Title != nil {
		task.Title = *updateReq.Title
	}

	if updateReq.Description != nil {
		task.Description = *updateReq.Description
	}

	if updateReq.Status != nil {
		task.Status = *updateReq.Status
	}
	if updateReq.Assignee != nil {
		task.AssigneeID = updateReq.Assignee
	}
	return task
}

func AssignedToEntity(assReq TaskRequest.AssignTaskRequest) entities.Task{
   
return entities.Task{
	AssigneeID: &assReq.AssigneeID,
}
}
