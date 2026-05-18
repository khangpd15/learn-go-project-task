package services

// import (
// 	"testing"

// 	"task_api/internal/entities"

// 	"github.com/stretchr/testify/assert"
// )

// type MockTaskRepository struct {
// 	tasks []entities.Task
// }

// func (m *MockTaskRepository) GetAllTasks() []entities.Task {
// 	return m.tasks
// }

// func (m *MockTaskRepository) GetTaskById(id int) *entities.Task {
// 	for _, task := range m.tasks {
// 		if task.ID == id {
// 			return &task
// 		}
// 	}
// 	return nil
// }

// func (m *MockTaskRepository) CreateTask(task entities.Task) entities.Task {
// 	task.ID = len(m.tasks) + 1
// 	m.tasks = append(m.tasks, task)
// 	return task
// }

// func (m *MockTaskRepository) UpdateTask(id int, updateTask entities.Task) *entities.Task {
// 	for i := range m.tasks {
// 		if m.tasks[i].ID == id {
// 			updateTask.ID = id
// 			m.tasks[i] = updateTask
// 			return &m.tasks[i]
// 		}
// 	}
// 	return nil
// }

// func (m *MockTaskRepository) DeleteTask(id int) bool {
// 	for i, task := range m.tasks {
// 		if task.ID == id {
// 			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
// 			return true
// 		}
// 	}
// 	return false
// }

// func TestCreateTaskSuccess(t *testing.T) {
// 	mockRepo := &MockTaskRepository{}
// 	taskService := NewTaskService(mockRepo)

// 	task := entities.Task{
// 		Title:  "Learn Go",
// 		Status: "TODO",
// 	}

// 	result, err := taskService.CreateTask(task)

// 	assert.NoError(t, err)
// 	assert.Equal(t, "Learn Go", result.Title)
// 	assert.Equal(t, "TODO", result.Status)
// 	assert.Equal(t, 1, result.ID)
// }

// func TestCreateTaskInvalidStatus(t *testing.T) {
// 	mockRepo := &MockTaskRepository{}
// 	taskService := NewTaskService(mockRepo)

// 	task := entities.Task{
// 		Title:  "Learn Go",
// 		Status: "INVALID",
// 	}

// 	result, err := taskService.CreateTask(task)

// 	assert.Error(t, err)
// 	assert.Equal(t, "invalid status", err.Error())
// 	assert.Equal(t, entities.Task{}, result)
// }

// func TestUpdateTaskSuccess(t *testing.T) {
// 	mockRepo := &MockTaskRepository{
// 		tasks: []entities.Task{
// 			{
// 				ID:     1,
// 				Title:  "Old Task",
// 				Status: "TODO",
// 			},
// 		},
// 	}
// 	taskService := NewTaskService(mockRepo)

// 	updateTask := entities.Task{
// 		Title:  "Updated Task",
// 		Status: "DONE",
// 	}

// 	result, err := taskService.UpdateTask(1, updateTask)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, result)
// 	assert.Equal(t, 1, result.ID)
// 	assert.Equal(t, "Updated Task", result.Title)
// 	assert.Equal(t, "DONE", result.Status)
// }

// func TestUpdateTaskInvalidID(t *testing.T) {
// 	mockRepo := &MockTaskRepository{}
// 	taskService := NewTaskService(mockRepo)

// 	updateTask := entities.Task{
// 		Title:  "Updated Task",
// 		Status: "DONE",
// 	}

// 	result, err := taskService.UpdateTask(0, updateTask)

// 	assert.Error(t, err)
// 	assert.Nil(t, result)
// 	assert.Equal(t, "invalid id", err.Error())
// }

// func TestUpdateTaskInvalidStatus(t *testing.T) {
// 	mockRepo := &MockTaskRepository{}
// 	taskService := NewTaskService(mockRepo)

// 	updateTask := entities.Task{
// 		Title:  "Updated Task",
// 		Status: "INVALID",
// 	}

// 	result, err := taskService.UpdateTask(1, updateTask)

// 	assert.Error(t, err)
// 	assert.Nil(t, result)
// 	assert.Equal(t, "invalid status", err.Error())
// }

// func TestUpdateTaskNotFound(t *testing.T) {
// 	mockRepo := &MockTaskRepository{}
// 	taskService := NewTaskService(mockRepo)

// 	updateTask := entities.Task{
// 		Title:  "Updated Task",
// 		Status: "DONE",
// 	}

// 	result, err := taskService.UpdateTask(999, updateTask)

// 	assert.Error(t, err)
// 	assert.Nil(t, result)
// 	assert.Equal(t, "task not found", err.Error())
// }

// func TestDeleteTaskSuccess(t *testing.T) {
// 	mockRepo := &MockTaskRepository{
// 		tasks: []entities.Task{
// 			{
// 				ID:     1,
// 				Title:  "Task Delete",
// 				Status: "TODO",
// 			},
// 		},
// 	}
// 	taskService := NewTaskService(mockRepo)

// 	err := taskService.DeleteTask(1)

// 	assert.NoError(t, err)
// }

// func TestDeleteTaskInvalidID(t *testing.T) {
// 	mockRepo := &MockTaskRepository{}
// 	taskService := NewTaskService(mockRepo)

// 	err := taskService.DeleteTask(0)

// 	assert.Error(t, err)
// 	assert.Equal(t, "invalid id", err.Error())
// }

// func TestDeleteTaskNotFound(t *testing.T) {
// 	mockRepo := &MockTaskRepository{}
// 	taskService := NewTaskService(mockRepo)

// 	err := taskService.DeleteTask(999)

// 	assert.Error(t, err)
// 	assert.Equal(t, "task not found", err.Error())
// }

