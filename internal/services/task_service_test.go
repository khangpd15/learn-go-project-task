package services

import (
	"errors"
	"testing"

	"task_api/internal/entities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTaskRepository struct {
	getAllTasksFn        func() ([]entities.Task, error)
	getTaskByIDFn        func(id int) (*entities.Task, error)
	createTaskFn         func(task entities.Task) (entities.Task, error)
	updateTaskFn         func(id int, task entities.Task) (*entities.Task, error)
	deleteTaskFn         func(id int) error
	getTaskListByProject func(projectID int) ([]entities.Task, error)
}

func (m *mockTaskRepository) GetAllTasks() ([]entities.Task, error) {
	if m.getAllTasksFn != nil {
		return m.getAllTasksFn()
	}
	return nil, nil
}

func (m *mockTaskRepository) GetTaskById(id int) (*entities.Task, error) {
	if m.getTaskByIDFn != nil {
		return m.getTaskByIDFn(id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTaskRepository) CreateTask(task entities.Task) (entities.Task, error) {
	if m.createTaskFn != nil {
		return m.createTaskFn(task)
	}
	return entities.Task{}, errors.New("not implemented")
}

func (m *mockTaskRepository) UpdateTask(id int, task entities.Task) (*entities.Task, error) {
	if m.updateTaskFn != nil {
		return m.updateTaskFn(id, task)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTaskRepository) DeleteTask(id int) error {
	if m.deleteTaskFn != nil {
		return m.deleteTaskFn(id)
	}
	return nil
}

func (m *mockTaskRepository) GetTaskListByProjectID(projectID int) ([]entities.Task, error) {
	if m.getTaskListByProject != nil {
		return m.getTaskListByProject(projectID)
	}
	return nil, nil
}

type mockProjectRepository struct {
	listByOwnerFn func(ownerID int) ([]entities.Project, error)
	listAllFn     func() ([]entities.Project, error)
	getByIDFn     func(projectID int) (entities.Project, error)
	updateFn      func(projectID int, name string, description string) error
	deleteFn      func(project entities.Project) error
}

func (m *mockProjectRepository) ListProjectByOwner(ownerID int) ([]entities.Project, error) {
	if m.listByOwnerFn != nil {
		return m.listByOwnerFn(ownerID)
	}
	return nil, nil
}

func (m *mockProjectRepository) ListAllProjects() ([]entities.Project, error) {
	if m.listAllFn != nil {
		return m.listAllFn()
	}
	return nil, nil
}

func (m *mockProjectRepository) GetProjectByID(projectID int) (entities.Project, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(projectID)
	}
	return entities.Project{}, errors.New("not implemented")
}

func (m *mockProjectRepository) UpdateProject(projectID int, name string, description string) error {
	if m.updateFn != nil {
		return m.updateFn(projectID, name, description)
	}
	return nil
}

func (m *mockProjectRepository) DeleteProject(project entities.Project) error {
	if m.deleteFn != nil {
		return m.deleteFn(project)
	}
	return nil
}

func TestTaskService_GetAllTasks_Success(t *testing.T) {
	taskRepo := &mockTaskRepository{
		getTaskListByProject: func(projectID int) ([]entities.Task, error) {
			switch projectID {
			case 1:
				return []entities.Task{{ID: 11, ProjectID: 1, Title: "Task 1", Status: "TODO"}}, nil
			case 2:
				return []entities.Task{{ID: 22, ProjectID: 2, Title: "Task 2", Status: "DONE"}}, nil
			default:
				return nil, nil
			}
		},
	}
	projectRepo := &mockProjectRepository{
		listByOwnerFn: func(ownerID int) ([]entities.Project, error) {
			require.Equal(t, 99, ownerID)
			return []entities.Project{{ID: 1, OwnerID: 99}, {ID: 2, OwnerID: 99}}, nil
		},
	}

	service := NewTaskService(taskRepo, projectRepo)
	tasks, err := service.GetAllTasks(99)

	require.NoError(t, err)
	require.Len(t, tasks, 2)
	assert.Equal(t, 11, tasks[0].ID)
	assert.Equal(t, 22, tasks[1].ID)
}

func TestTaskService_GetTaskById_Success(t *testing.T) {
	taskRepo := &mockTaskRepository{
		getTaskByIDFn: func(id int) (*entities.Task, error) {
			return &entities.Task{ID: id, ProjectID: 7, Title: "Task", Status: "TODO"}, nil
		},
	}
	projectRepo := &mockProjectRepository{
		getByIDFn: func(projectID int) (entities.Project, error) {
			return entities.Project{ID: projectID, OwnerID: 12}, nil
		},
	}

	service := NewTaskService(taskRepo, projectRepo)
	task, err := service.GetTaskById(12, 5)

	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, 5, task.ID)
	assert.Equal(t, 7, task.ProjectID)
}

func TestTaskService_GetTaskById_Forbidden(t *testing.T) {
	taskRepo := &mockTaskRepository{
		getTaskByIDFn: func(id int) (*entities.Task, error) {
			return &entities.Task{ID: id, ProjectID: 7, Title: "Task", Status: "TODO"}, nil
		},
	}
	projectRepo := &mockProjectRepository{
		getByIDFn: func(projectID int) (entities.Project, error) {
			return entities.Project{ID: projectID, OwnerID: 99}, nil
		},
	}

	service := NewTaskService(taskRepo, projectRepo)
	task, err := service.GetTaskById(12, 5)

	assert.Error(t, err)
	assert.Equal(t, "forbidden", err.Error())
	assert.Nil(t, task)
}

func TestTaskService_CreateTask_Success(t *testing.T) {
	var receivedTask entities.Task
	taskRepo := &mockTaskRepository{
		createTaskFn: func(task entities.Task) (entities.Task, error) {
			receivedTask = task
			task.ID = 100
			return task, nil
		},
	}
	projectRepo := &mockProjectRepository{
		getByIDFn: func(projectID int) (entities.Project, error) {
			return entities.Project{ID: projectID, OwnerID: 7}, nil
		},
	}

	service := NewTaskService(taskRepo, projectRepo)
	result, err := service.CreateTask(7, entities.Task{ProjectID: 3, Title: "New task", Status: "TODO"})

	require.NoError(t, err)
	require.Equal(t, 100, result.ID)
	require.NotNil(t, receivedTask.AssigneeID)
	assert.Equal(t, 7, *receivedTask.AssigneeID)
	assert.Equal(t, "TODO", receivedTask.Status)
}

func TestTaskService_CreateTask_InvalidStatus(t *testing.T) {
	taskRepo := &mockTaskRepository{}
	projectRepo := &mockProjectRepository{
		getByIDFn: func(projectID int) (entities.Project, error) {
			return entities.Project{ID: projectID, OwnerID: 7}, nil
		},
	}

	service := NewTaskService(taskRepo, projectRepo)
	result, err := service.CreateTask(7, entities.Task{ProjectID: 3, Title: "New task", Status: "invalid"})

	assert.Error(t, err)
	assert.Equal(t, "invalid status", err.Error())
	assert.Equal(t, entities.Task{}, result)
}

func TestTaskService_CreateTask_ForbiddenProjectOwner(t *testing.T) {
	taskRepo := &mockTaskRepository{}
	projectRepo := &mockProjectRepository{
		getByIDFn: func(projectID int) (entities.Project, error) {
			return entities.Project{ID: projectID, OwnerID: 99}, nil
		},
	}

	service := NewTaskService(taskRepo, projectRepo)
	result, err := service.CreateTask(7, entities.Task{ProjectID: 3, Title: "New task", Status: "TODO"})

	assert.Error(t, err)
	assert.Equal(t, "unauthorized to create task in this project", err.Error())
	assert.Equal(t, entities.Task{}, result)
}

func TestTaskService_UpdateTask_Success(t *testing.T) {
	var receivedTask entities.Task
	taskRepo := &mockTaskRepository{
		getTaskByIDFn: func(id int) (*entities.Task, error) {
			return &entities.Task{ID: id, ProjectID: 8, Title: "Old", Status: "TODO"}, nil
		},
		updateTaskFn: func(id int, task entities.Task) (*entities.Task, error) {
			receivedTask = task
			task.ID = id
			return &task, nil
		},
	}
	projectRepo := &mockProjectRepository{
		getByIDFn: func(projectID int) (entities.Project, error) {
			return entities.Project{ID: projectID, OwnerID: 7}, nil
		},
	}

	service := NewTaskService(taskRepo, projectRepo)
	result, err := service.UpdateTask(5, 7, entities.Task{Title: "Updated", Status: "done"})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 5, result.ID)
	assert.Equal(t, "DONE", receivedTask.Status)
}

func TestTaskService_UpdateTask_InvalidStatus(t *testing.T) {
	taskRepo := &mockTaskRepository{
		getTaskByIDFn: func(id int) (*entities.Task, error) {
			return &entities.Task{ID: id, ProjectID: 8, Title: "Old", Status: "TODO"}, nil
		},
	}
	projectRepo := &mockProjectRepository{
		getByIDFn: func(projectID int) (entities.Project, error) {
			return entities.Project{ID: projectID, OwnerID: 7}, nil
		},
	}

	service := NewTaskService(taskRepo, projectRepo)
	result, err := service.UpdateTask(5, 7, entities.Task{Title: "Updated", Status: "bad"})

	assert.Error(t, err)
	assert.Equal(t, "invalid status", err.Error())
	assert.Nil(t, result)
}

func TestTaskService_UpdateTask_Forbidden(t *testing.T) {
	taskRepo := &mockTaskRepository{
		getTaskByIDFn: func(id int) (*entities.Task, error) {
			return &entities.Task{ID: id, ProjectID: 8, Title: "Old", Status: "TODO"}, nil
		},
	}
	projectRepo := &mockProjectRepository{
		getByIDFn: func(projectID int) (entities.Project, error) {
			return entities.Project{ID: projectID, OwnerID: 99}, nil
		},
	}

	service := NewTaskService(taskRepo, projectRepo)
	result, err := service.UpdateTask(5, 7, entities.Task{Title: "Updated", Status: "TODO"})

	assert.Error(t, err)
	assert.Equal(t, "unauthorized to update this task", err.Error())
	assert.Nil(t, result)
}

func TestTaskService_DeleteTask_Success(t *testing.T) {
	taskRepo := &mockTaskRepository{
		getTaskByIDFn: func(id int) (*entities.Task, error) {
			return &entities.Task{ID: id, ProjectID: 4, Title: "Delete me", Status: "TODO"}, nil
		},
		deleteTaskFn: func(id int) error {
			return nil
		},
	}
	projectRepo := &mockProjectRepository{
		getByIDFn: func(projectID int) (entities.Project, error) {
			return entities.Project{ID: projectID, OwnerID: 7}, nil
		},
	}

	service := NewTaskService(taskRepo, projectRepo)
	err := service.DeleteTask(5, 7)

	require.NoError(t, err)
}

func TestTaskService_DeleteTask_Forbidden(t *testing.T) {
	taskRepo := &mockTaskRepository{
		getTaskByIDFn: func(id int) (*entities.Task, error) {
			return &entities.Task{ID: id, ProjectID: 4, Title: "Delete me", Status: "TODO"}, nil
		},
	}
	projectRepo := &mockProjectRepository{
		getByIDFn: func(projectID int) (entities.Project, error) {
			return entities.Project{ID: projectID, OwnerID: 99}, nil
		},
	}

	service := NewTaskService(taskRepo, projectRepo)
	err := service.DeleteTask(5, 7)

	assert.Error(t, err)
	assert.Equal(t, "unauthorized to delete this task", err.Error())
}
