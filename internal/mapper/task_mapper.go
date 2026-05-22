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
		
	}
}
func UpdateTaskRequestToTaskEntity(updateReq TaskRequest.UpdateTaskRequest) entities.Task {
	return entities.Task{
		Title:       updateReq.Title,
		Description: updateReq.Description,
		Status:      updateReq.Status,
	}
}

