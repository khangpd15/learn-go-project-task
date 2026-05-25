package services

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"task_api/internal/cache"
	requestDTO "task_api/internal/dto/request/task"
	"task_api/internal/entities"
	"task_api/internal/repositories"
	"task_api/internal/validation"
	"time"
)

type TaskService struct {
	taskRepo    repositories.TaskRepositoryInterface
	projectRepo repositories.ProjectRepositoryInterface
	userRepo    repositories.UserRepositoryInterface
	cache       cache.Cache
}

func NewTaskService(
	taskRepo repositories.TaskRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	cacheClient cache.Cache,
) *TaskService {
	return &TaskService{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		userRepo:    userRepo,
		cache:       cacheClient,
	}
}

func (s *TaskService) GetAllTasks(currentUserID int, ctx context.Context) ([]entities.Task, error) {
	start := time.Now()
	key := cache.TaskListByUserKey(currentUserID)
	var allTasks []entities.Task
	cachedData, err := s.cache.Get(ctx, key)
	if err == nil && cachedData != "" {
		var cachedTasks []entities.Task
		if err := json.Unmarshal([]byte(cachedData), &cachedTasks); err == nil {
			log.Println("Cache Hit", key)
			log.Println("Duration:", time.Since(start))
			allTasks = cachedTasks
		}
	}
	if allTasks == nil {
		log.Println("Cache Miss", key)
		tasks, err := s.taskRepo.GetAllTasksByUserID(currentUserID)
		if err != nil {
			return nil, err
		}
		allTasks = tasks
		data, err := json.Marshal(allTasks)
		if err == nil {
			if err := s.cache.Set(ctx, key, data, 5*time.Minute); err != nil {
				log.Println("Failed to set cache:", err)
			} else {
				log.Println("Cache saved:", key)
			}
		}
	}
	

	return allTasks, nil
}

func (s *TaskService) GetTaskById(currentUserID int, ctx context.Context, id int) (*entities.Task, error) {
	if !validation.IsValidId(id) {
		return nil, ErrInvalidTaskID
	}

	start := time.Now()
	key := cache.TaskKey(id)
	var task *entities.Task
	cachedData, err := s.cache.Get(ctx, key)
	if err == nil && cachedData != "" {
		var cachedTask entities.Task
		if err := json.Unmarshal([]byte(cachedData), &cachedTask); err == nil {
			log.Println("Cache Hit", key)
			log.Println("Duration:", time.Since(start))
			task = &cachedTask
		}
	}
	if task == nil {
		log.Println("Cache Miss", key)

		dbTask, err := s.taskRepo.GetTaskById(id)
		if err != nil {
			return nil, ErrTaskNotFound
		}

		task = dbTask
		data, err := json.Marshal(task)
		if err == nil {
			if err := s.cache.Set(ctx, key, data, 5*time.Minute); err != nil {
				log.Println("Failed to set cache:", err)
			} else {
				log.Println("Cache saved:", key)
			}
		}
	}

	project, err := s.projectRepo.GetProjectByID(task.ProjectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	isOwner := project.OwnerID == currentUserID

	isAssignee := false

	if task.AssigneeID != nil {
		isAssignee = *task.AssigneeID == currentUserID
	}

	if !isOwner && !isAssignee {
		return nil, ErrForbidden
	}
	return task, nil
}

func (s *TaskService) CreateTask(currentUserID int, task entities.Task) (entities.Task, error) {
	if !validation.IsValidId(task.ProjectID) {
		return entities.Task{}, ErrInvalidProjectID
	}

	project, err := s.projectRepo.GetProjectByID(task.ProjectID)
	if err != nil {
		return entities.Task{}, ErrProjectNotFound
	}

	if project.OwnerID != currentUserID {
		return entities.Task{}, ErrForbidden
	}

	task.Status = strings.ToUpper(task.Status)
	if !validation.IsValidStatus(task.Status) {
		return entities.Task{}, ErrInvalidStatus
	}

	task.AssigneeID = nil

	created, err := s.taskRepo.CreateTask(task)
	if err != nil {
		return entities.Task{}, err
	}
	// Invalidate caches related to tasks
	ctx := context.Background()
	_ = s.cache.Delete(ctx, cache.TaskKey(created.ID))
	return created, nil
}

func (s *TaskService) UpdateTask(id int, currentUserID int, req requestDTO.UpdateTaskRequest) (*entities.Task, error) {
	if !validation.IsValidId(id) {
		return nil, ErrInvalidTaskID
	}

	task, err := s.taskRepo.GetTaskById(id)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	project, err := s.projectRepo.GetProjectByID(task.ProjectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	if project.OwnerID != currentUserID {
		return nil, ErrForbidden
	}

	updateTask := entities.Task{}

	if req.Title != nil {
		updateTask.Title = *req.Title
	}

	if req.Description != nil {
		updateTask.Description = *req.Description
	}

	if req.Status != nil {
		status := strings.ToUpper(*req.Status)
		if !validation.IsValidStatus(status) {
			return nil, ErrInvalidStatus
		}
		updateTask.Status = status
	}

	updated, err := s.taskRepo.UpdateTask(id, updateTask)
	if err != nil {
		return nil, err
	}
	// Invalidate caches
	ctx := context.Background()
	_ = s.cache.Delete(ctx, cache.TaskKey(id))

	return updated, nil
}

func (s *TaskService) DeleteTask(id int, currentUserID int) error {
	if !validation.IsValidId(id) {
		return ErrInvalidTaskID
	}

	task, err := s.taskRepo.GetTaskById(id)
	if err != nil {
		return ErrTaskNotFound
	}

	project, err := s.projectRepo.GetProjectByID(task.ProjectID)
	if err != nil {
		return ErrProjectNotFound
	}

	if project.OwnerID != currentUserID {
		return ErrForbidden
	}

	err = s.taskRepo.DeleteTask(id)
	if err != nil {
		return err
	}
	// Invalidate cache
	ctx := context.Background()
	_ = s.cache.Delete(ctx, cache.TaskKey(id))
	return nil
}
