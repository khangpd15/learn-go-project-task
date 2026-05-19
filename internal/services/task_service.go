package services

import (
	"errors"
	"task_api/internal/entities"
	"task_api/internal/repositories"
	"task_api/internal/validation"
)

type TaskService struct {
	taskRepo repositories.TaskRepositoryInterface
}

func NewTaskService(repo repositories.TaskRepositoryInterface) *TaskService {
	return &TaskService{taskRepo: repo}
}



func (s *TaskService) GetAllTasks() ([]entities.Task, error) {
	return s.taskRepo.GetAllTasks()
}

func (s *TaskService) GetTaskById(id int) (*entities.Task, error) {
	if !validation.IsValidId(id) {
		return nil, errors.New("invalid id")
	}

	return s.taskRepo.GetTaskById(id)
}
func (s *TaskService) CreateTask(task entities.Task) (entities.Task, error) {
	if !validation.IsValidStatus(task.Status) {
		return entities.Task{}, errors.New("invalid status")
	}

	return s.taskRepo.CreateTask(task)
}

func (s *TaskService) UpdateTask(id int, updateTask entities.Task) (*entities.Task, error) {
	if !validation.IsValidId(id) {
		return nil, errors.New("invalid id")
	}

	if !validation.IsValidStatus(updateTask.Status) {
		return nil, errors.New("invalid status")
	}

	return s.taskRepo.UpdateTask(id, updateTask)
}

func (s *TaskService) DeleteTask(id int) error {
	if !validation.IsValidId(id) {
		return errors.New("invalid id")
	}

	return s.taskRepo.DeleteTask(id)
}
