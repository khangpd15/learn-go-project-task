package services

import (
	"errors"
	"task_api/internal/entities"
	"task_api/internal/repositories"
	"task_api/internal/validation"
)

var (
	ErrInvalidProjectID      = errors.New("invalid project id")
	ErrInvalidProjectOwnerId = errors.New("invalid project owner id")
	ErrProjectNotFound       = errors.New("project not found")
	ErrForbidden             = errors.New("forbidden")
)

type ProjectService struct {
	projectRepository repositories.ProjectRepositoryInterface
}

func NewProjectService(projectRepo repositories.ProjectRepositoryInterface) *ProjectService {
	return &ProjectService{
		projectRepository: projectRepo,
	}
}
func (s *ProjectService) ListProjectsByOwner(ownerID int) ([]entities.Project, error) {

	if !validation.IsValidProjectOwnerId(ownerID) {
		return nil, ErrInvalidProjectOwnerId
	}

	projects, err := s.projectRepository.ListProjectByOwner(ownerID)
	if err != nil {
		return nil, err
	}

	return projects, nil
}
func (s *ProjectService) ListAllProjects() ([]entities.Project, error) {
	projects, err := s.projectRepository.ListAllProjects()
	if err != nil {
		return nil, err
	}
	return projects, nil
}
func (s *ProjectService) GetProjectByID(projectID int) (entities.Project, error) {
	if !validation.IsValidProjectId(projectID) {
		return entities.Project{}, ErrInvalidProjectID
	}
	project, err := s.projectRepository.GetProjectByID(projectID)
	if err != nil {
		return entities.Project{}, ErrProjectNotFound
	}
	return project, nil
}

func (s *ProjectService) UpdateProject(currentUserID int, projectID int, name *string, description *string) error {
	if !validation.IsValidProjectId(projectID) {
		return ErrInvalidProjectID
	}
	project, err := s.projectRepository.GetProjectByID(projectID)
	if err != nil {
		return ErrProjectNotFound
	}

	if project.OwnerID != currentUserID {
		return ErrForbidden
	}

	if name != nil {
		project.Name = *name
	}

	if description != nil {
		project.Description = *description
	}

	return s.projectRepository.UpdateProject(projectID, project.Name, project.Description)
}

func (s *ProjectService) DeleteProject(currentUserID int, projectID int) error {
	if !validation.IsValidProjectId(projectID) {
		return ErrInvalidProjectID
	}

	project, err := s.projectRepository.GetProjectByID(projectID)
	if err != nil {
		return ErrProjectNotFound
	}

	if project.OwnerID != currentUserID {
		return ErrForbidden
	}

	return s.projectRepository.DeleteProject(project)
}
