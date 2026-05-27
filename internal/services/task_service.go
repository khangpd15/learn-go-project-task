package services

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"task_api/internal/cache"
	requestDTO "task_api/internal/dto/request/task"
	"task_api/internal/entities"
	"task_api/internal/events"
	"task_api/internal/jobs"
	"task_api/internal/queue"
	"task_api/internal/repositories"
	"task_api/internal/validation"
	"time"
)

type TaskService struct {
	taskRepo       repositories.TaskRepositoryInterface
	projectRepo    repositories.ProjectRepositoryInterface
	userRepo       repositories.UserRepositoryInterface
	cache          cache.Cache
	queue          queue.Queue
	eventPublisher events.Publisher
}

func NewTaskService(
	taskRepo repositories.TaskRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	cacheClient cache.Cache,
	queue queue.Queue,
	eventPublisher events.Publisher,
) *TaskService {
	return &TaskService{
		taskRepo:       taskRepo,
		projectRepo:    projectRepo,
		userRepo:       userRepo,
		cache:          cacheClient,
		queue:          queue,
		eventPublisher: eventPublisher,
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

func (s *TaskService) CreateTask(currentUserID int, ctx context.Context, task entities.Task) (entities.Task, error) {
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
	_ = s.cache.Delete(ctx, cache.TaskListByUserKey(currentUserID))
	_ = s.eventPublisher.Publish(ctx, events.Event{
		Type:      events.EventTaskCreated,
		UserID:    currentUserID,
		ProjectID: created.ProjectID,
		TaskID:    created.ID,
		Data: map[string]interface{}{
			"id":          created.ID,
			"project_id":  created.ProjectID,
			"title":       created.Title,
			"description": created.Description,
			"status":      created.Status,
			"assignee_id": created.AssigneeID,
			"created_at":  created.CreatedAt,
		},
	})
	return created, nil
}

func (s *TaskService) UpdateTask(id int, currentUserID int, ctx context.Context, req requestDTO.UpdateTaskRequest) (*entities.Task, error) {
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

	updates := map[string]interface{}{}

	if req.Title != nil {
		updates["title"] = *req.Title
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.Status != nil {
		status := strings.ToUpper(*req.Status)
		if !validation.IsValidStatus(status) {
			return nil, ErrInvalidStatus
		}
		updates["status"] = status
	}

	assigneeChanged := false
	var newAssigneeID int
	if req.Assignee != nil {

		if *req.Assignee == 0 {

			updates["assignee_id"] = nil
			log.Println("task dang chon lai nguoi lam")

		} else {

			if !validation.IsValidId(*req.Assignee) {
				return nil, ErrInvalidAssigneeID
			}

			_, err := s.userRepo.GetUserByID(*req.Assignee)
			if err != nil {
				return nil, ErrUserNotFound
			}

			updates["assignee_id"] = *req.Assignee

			assigneeChanged = true
			newAssigneeID = *req.Assignee
		}
	}

	updated, err := s.taskRepo.UpdateTask(id, updates)
	if err != nil {
		return nil, err
	}

	if assigneeChanged {
		job := jobs.NotificationJob{
			TaskID:     id,
			AssigneeID: newAssigneeID,
			Message:    "You have been assigned a new task",
			RetryCount: 0,
		}

		payload, err := json.Marshal(job)
		if err != nil {
			log.Println("failed to marshal notification job:", err)
		} else {
			if err := s.queue.Enqueue(ctx, jobs.NotificationQueueName, payload); err != nil {
				log.Println("failed to enqueue notification job:", err)
			}
		}
	}
	// Invalidate caches

	_ = s.cache.Delete(ctx, cache.TaskKey(id))
	_ = s.cache.Delete(ctx, cache.TaskListByUserKey(currentUserID))
	if assigneeChanged {
		_ = s.cache.Delete(ctx, cache.TaskListByUserKey(newAssigneeID))
	}

	return updated, nil
}

func (s *TaskService) DeleteTask(id int, currentUserID int, ctx context.Context) error {
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

	_ = s.cache.Delete(ctx, cache.TaskKey(id))
	_ = s.cache.Delete(ctx, cache.TaskListByUserKey(currentUserID))
	return nil
}

func (s *TaskService) AssignTask(id int, currentUserID int, assigneeID int, ctx context.Context) (*entities.Task, error) {
	if !validation.IsValidId(id) {
		return nil, ErrInvalidTaskID
	}
	if !validation.IsValidIdAssignTask(assigneeID) {
		return nil, ErrInvalidAssigneeID
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
	user, err := s.userRepo.GetUserByID(assigneeID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	println("Assigned task", id, "to user", user.ID)
	task.AssigneeID = &assigneeID

	assign, err := s.taskRepo.AssignTask(id, *task.AssigneeID)
	if err != nil {
		return nil, err
	}
	_ = s.cache.Delete(ctx, cache.TaskKey(id))
	_ = s.cache.Delete(ctx, cache.TaskListByUserKey(currentUserID))
	_ = s.cache.Delete(ctx, cache.TaskListByUserKey(assigneeID))

	job := jobs.NotificationJob{
		TaskID:     id,
		AssigneeID: assigneeID,
		Message:    "You have been assigned a new task",
		RetryCount: 0,
	}
	payload, err := json.Marshal(job)
	if err != nil {
		log.Println("Failed to marshal job:", err)
	} else {
		if err := s.queue.Enqueue(ctx, jobs.NotificationQueueName, payload); err != nil {
			log.Println("Failed to enqueue job:", err)
		}
	}

	return assign, nil
}
func (s *TaskService) UnassignTask(ctx context.Context, taskID int, currentUserID int) (*entities.Task, error) {
	if !validation.IsValidId(taskID) {
		return nil, ErrInvalidTaskID
	}

	task, err := s.taskRepo.GetTaskById(taskID)
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

	updated, err := s.taskRepo.UnassignTask(taskID)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Delete(ctx, cache.TaskKey(taskID))
	_ = s.cache.Delete(ctx, cache.TaskListByUserKey(currentUserID))

	return updated, nil
}
