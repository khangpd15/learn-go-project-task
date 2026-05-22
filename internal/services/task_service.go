package services

import (
	"errors"
	"task_api/internal/entities"
	"task_api/internal/repositories"
	"task_api/internal/validation"
)

type TaskService struct {
	taskRepo repositories.TaskRepositoryInterface
	projectRepo repositories.ProjectRepositoryInterface
}

func NewTaskService(repo repositories.TaskRepositoryInterface, projectRepo repositories.ProjectRepositoryInterface) *TaskService {
	return &TaskService{
		taskRepo: repo,
		projectRepo: projectRepo,
	}
}



func (s *TaskService) GetAllTasks(currentUserID int) ([]entities.Task, error) {
	projects, err := s.projectRepo.ListProjectByOwner(currentUserID)
	if err != nil {
		return nil, err
	}
	var allTasks []entities.Task
	for _, project := range projects {
		tasks, err := s.taskRepo.GetTaskListByProjectID(project.ID)
		if err != nil {
			return nil, err
		}
		allTasks = append(allTasks, tasks...)
	}
	return allTasks, nil
}


func (s *TaskService) GetTaskById(currentUserID, id int) (*entities.Task, error) {
	if !validation.IsValidId(id) {
		return nil, errors.New("invalid id")
	}

	getTask, err := s.taskRepo.GetTaskById(id)
	if err != nil {
		return nil, errors.New("task not found")
	}

	// Validate ownership: check if task belongs to current user's project
	project, err := s.projectRepo.GetProjectByID(getTask.ProjectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	if project.OwnerID != currentUserID {
		return nil, errors.New("forbidden")
	}

	return getTask, nil
}

func (s *TaskService) CreateTask(currentUserID int,task entities.Task) (entities.Task, error) {
	project, err := s.projectRepo.GetProjectByID(task.ProjectID)
	if err != nil {
		return entities.Task{}, err
	}
	if project.OwnerID != currentUserID {
		return entities.Task{}, errors.New("unauthorized to create task in this project")
	}
	if !validation.IsValidStatus(task.Status) {
		return entities.Task{}, errors.New("invalid status")
	}
	task.AssigneeID = &currentUserID
	return s.taskRepo.CreateTask(task)
}

func (s *TaskService) UpdateTask(id int,currentUserID int, updateTask entities.Task) (*entities.Task, error) {
	task, err := s.taskRepo.GetTaskById(id)
	if err != nil {
		return nil, err
	}
	project, err := s.projectRepo.GetProjectByID(task.ProjectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if project.OwnerID != currentUserID {
		return nil, errors.New("unauthorized to update this task")
	}
	if !validation.IsValidId(id) {
		return nil, errors.New("invalid id")
	}

	if !validation.IsValidStatus(updateTask.Status) {
		return nil, errors.New("invalid status")
	}

	return s.taskRepo.UpdateTask(id, updateTask)
}

func (s *TaskService) DeleteTask(id int,currentUserID int) error {
	task, err := s.taskRepo.GetTaskById(id)
	if err != nil {
		return err
	}
	project, err := s.projectRepo.GetProjectByID(task.ProjectID)
	if err != nil {
		return err
	}
	if project.OwnerID != currentUserID {
		return errors.New("unauthorized to delete this task")
	}
	if !validation.IsValidId(id) {
		return errors.New("invalid id")
	}

	return s.taskRepo.DeleteTask(id)
}
