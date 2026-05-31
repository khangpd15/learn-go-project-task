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
	taskRepo            repositories.TaskRepositoryInterface
	projectRepo         repositories.ProjectRepositoryInterface
	userRepo            repositories.UserRepositoryInterface
	cache               cache.Cache
	queue               queue.Queue
	eventPublisher      events.Publisher
	notificationService NotificationService
}

func NewTaskService(
	taskRepo repositories.TaskRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	cacheClient cache.Cache,
	queue queue.Queue,
	eventPublisher events.Publisher,
	notificationService NotificationService,
) *TaskService {
	return &TaskService{
		taskRepo:            taskRepo,
		projectRepo:         projectRepo,
		userRepo:            userRepo,
		cache:               cacheClient,
		queue:               queue,
		eventPublisher:      eventPublisher,
		notificationService: notificationService,
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

	if task.AssigneeID != nil {
		if *task.AssigneeID == 0 {
			task.AssigneeID = nil
		} else {
			if !validation.IsValidId(*task.AssigneeID) {
				return entities.Task{}, ErrInvalidAssigneeID
			}

			_, err := s.userRepo.GetUserByID(*task.AssigneeID)
			if err != nil {
				return entities.Task{}, ErrUserNotFound
			}
		}
	}

	created, err := s.taskRepo.CreateTask(task)
	if err != nil {
		return entities.Task{}, err
	}
    log.Println("[TASK] created task id:", created.ID)
	if created.AssigneeID != nil {
		notification := entities.Notification{
			TaskID:     created.ID,
			ProjectID:  created.ProjectID,
			SenderID:   currentUserID,
			ReceiverID: created.AssigneeID,
			Title:      "Create Task Notification",
			Type:       "TASK_CREATED",
			Message:    "Congratulations! You have a new task and ADMIN just created it for you",
	
		}

		s.createNotificationAndEnqueue(ctx, notification)
		
	}

	_ = s.cache.Delete(ctx, cache.TaskListByUserKey(currentUserID))

	if created.AssigneeID != nil {
		assigneeID := *created.AssigneeID

		_ = s.cache.Delete(ctx, cache.TaskListByUserKey(assigneeID))

		log.Println("[TASK] create publish to assignee:", assigneeID)

		_ = s.eventPublisher.Publish(ctx, events.Event{
			Type:      events.EventTaskCreated,
			UserID:    assigneeID,
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
				"message":     "You have been assigned a new task",
			},
		})
	}

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

	statusChanged := false
	if req.Status != nil {
		status := strings.ToUpper(*req.Status)
		if !validation.IsValidStatus(status) {
			return nil, ErrInvalidStatus
		}

		updates["status"] = status
		statusChanged = true
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
		notification := entities.Notification{
			TaskID:     updated.ID,
			ProjectID:  updated.ProjectID,
			SenderID:   currentUserID,
			ReceiverID: &newAssigneeID,
			Title:      "Task Assigned",
			Type:       "TASK_ASSIGNED",
			Message:    "You have been assigned a new task",
	
		}

		s.createNotificationAndEnqueue(ctx, notification)
	}

	_ = s.cache.Delete(ctx, cache.TaskKey(id))
	_ = s.cache.Delete(ctx, cache.TaskListByUserKey(currentUserID))

	if assigneeChanged {
		_ = s.cache.Delete(ctx, cache.TaskListByUserKey(newAssigneeID))

		_ = s.eventPublisher.Publish(ctx, events.Event{
			Type:      events.EventTaskAssigned,
			UserID:    newAssigneeID,
			ProjectID: updated.ProjectID,
			TaskID:    updated.ID,
			Data: map[string]interface{}{
				"message": "You have been assigned a new task",
				"task_id": updated.ID,
			},
		})
	}

	if statusChanged {
		_ = s.eventPublisher.Publish(ctx, events.Event{
			Type:      events.EventUpdateStatus,
			UserID:    project.OwnerID,
			ProjectID: updated.ProjectID,
			TaskID:    updated.ID,
			Data: map[string]interface{}{
				"message":   "Task status changed",
				"OldStatus": task.Status,
				"NewStatus": updated.Status,
				"task_id":   updated.ID,
			},
		})
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

	log.Println("Assigned task", id, "to user", user.ID)

	task.AssigneeID = &assigneeID

	assign, err := s.taskRepo.AssignTask(id, *task.AssigneeID)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Delete(ctx, cache.TaskKey(id))
	_ = s.cache.Delete(ctx, cache.TaskListByUserKey(currentUserID))
	_ = s.cache.Delete(ctx, cache.TaskListByUserKey(assigneeID))

	notification := entities.Notification{
		TaskID:     assign.ID,
		ProjectID:  assign.ProjectID,
		SenderID:   currentUserID,
		ReceiverID: &assigneeID,
		Title:      "Task Assigned",
		Type:       "TASK_ASSIGNED",
		Message:    "You have been assigned a new task",
	
	}

	s.createNotificationAndEnqueue(ctx, notification)

	_ = s.eventPublisher.Publish(ctx, events.Event{
		Type:      events.EventTaskAssigned,
		UserID:    assigneeID,
		ProjectID: assign.ProjectID,
		TaskID:    assign.ID,
		Data: map[string]interface{}{
			"message": "You have been assigned a new task",
		},
		
	})

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

func (s *TaskService) createNotificationAndEnqueue(
	ctx context.Context,
	notification entities.Notification,
) {
	createdNotification, err := s.notificationService.CreateNotification(notification)
	if err != nil {
		log.Println("[TASK] failed to create notification:", err)
		return
	}
	log.Println("[NOTIFICATION] id:", createdNotification.ID)

	job := jobs.NotificationJob{
		NotificationID: createdNotification.ID,
		RetryCount:     0,
	}

	payload, err := json.Marshal(job)
	if err != nil {
		log.Println("[TASK] failed to marshal notification job:", err)
		return
	}

	if err := s.queue.Enqueue(ctx, jobs.NotificationQueueName, payload); err != nil {
		log.Println("[TASK] failed to enqueue notification job:", err)
	}
}