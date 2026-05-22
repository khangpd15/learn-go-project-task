package repositories
import (
	"task_api/internal/entities"
	"gorm.io/gorm"
)

type TaskRepositoryInterface interface {
GetAllTasks() ([]entities.Task, error)
GetTaskById(id int) (*entities.Task, error)
CreateTask(task entities.Task) (entities.Task, error)
UpdateTask(id int, updatedTask entities.Task) (*entities.Task, error)
DeleteTask(id int)  error
GetTaskListByProjectID(projectID int) ([]entities.Task, error)

}

type TaskRepository struct {
	db *gorm.DB
}
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}
func (r *TaskRepository) GetAllTasks() ([]entities.Task, error) {
	var tasks []entities.Task
	err := r.db.
	Order("id asc").
    Find(&tasks).
    Error
	if err != nil {
		return nil, err
	}
	return tasks, nil

}
func (r *TaskRepository) GetTaskListByProjectID(projectID int) ([]entities.Task, error) {
	var tasks []entities.Task
	err := r.db.Where("project_id = ?", projectID).
	Order("id asc").
	Find(&tasks).
	Error	
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TaskRepository) GetTaskById(id int) (*entities.Task, error) {
	var task entities.Task
	err := r.db.First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) CreateTask(task entities.Task) (entities.Task, error) {
	err := r.db.Create(&task).Error
	if err != nil {
		return entities.Task{}, err
	}
	return task, nil
}

func (r *TaskRepository) UpdateTask(id int, task entities.Task) (*entities.Task, error) {
	var existingTask entities.Task

	err := r.db.
		First(&existingTask, id).
		Error

	if err != nil {
		return nil, err
	}

	existingTask.Title = task.Title
	existingTask.Description = task.Description
	existingTask.Status = task.Status
	existingTask.AssigneeID = task.AssigneeID

	err = r.db.
		Save(&existingTask).
		Error

	if err != nil {
		return nil, err
	}

	return &existingTask, nil
}
func (r *TaskRepository) DeleteTask(id int) error {
	err := r.db.
		Delete(&entities.Task{}, id).
		Error

	return err
}