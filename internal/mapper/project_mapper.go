package mapper 

import (
	RequestDTO "task_api/internal/dto/request/project"
	ResponseDTO "task_api/internal/dto/response/project"
	"task_api/internal/entities"
)

func UpdateProjectRequestToEntity(req RequestDTO.UpdateProjectRequest) entities.Project {
	return entities.Project{
		Name : *req.Name,
		Description : *req.Description,
	}
}

func ProjectToResponse(project entities.Project) ResponseDTO.ProjectResponse {
	return ResponseDTO.ProjectResponse{
		ID: project.ID,
		Name: project.Name,
		Description: project.Description,
	}
}
func ProjectsToResponses(projects []entities.Project) []ResponseDTO.ProjectResponse {
	responses := make([]ResponseDTO.ProjectResponse, len(projects))	
	for i, project := range projects {
		responses[i] = ProjectToResponse(project)
	}
	return responses
}