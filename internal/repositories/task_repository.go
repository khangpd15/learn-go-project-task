package repositories
import (
	"task_api/internal/entities"
	"gorm.io/gorm"
)

type TaskRepositoryInterface interface {
GetAllTasks() ([]entities.Task, error)
GetTaskById(id int) (*entities.Task, error)
CreateTask(task entities.Task) (entities.Task, error)
UpdateTask(id int, updates map[string]interface{}) (*entities.Task, error)
DeleteTask(id int)  error
GetTaskListByProjectID(projectID int) ([]entities.Task, error)
GetAllTasksByUserID(userID int) ([]entities.Task, error)
AssignTask(id int, assigneeID int) (*entities.Task, error)
UnassignTask(id int) (*entities.Task, error)
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
func (r *TaskRepository) GetAllTasksByUserID(userID int) ([]entities.Task, error) {
	var tasks []entities.Task

	err := r.db.
		Joins("JOIN projects ON projects.id = tasks.project_id").
		Where("projects.owner_id = ? OR tasks.assignee_id = ?", userID, userID).
		Order("tasks.id ASC").
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

func (r *TaskRepository) AssignTask(id int, assigneeID int) (*entities.Task, error) {
	var task entities.Task

	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&task).Update("assignee_id", assigneeID).Error; err != nil {
		return nil, err
	}

	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *TaskRepository) UnassignTask(id int) (*entities.Task, error) {
	var task entities.Task

	err := r.db.First(&task, id).Error
	if err != nil {
		return nil, err
	}

	task.AssigneeID = nil

	err = r.db.Save(&task).Error
	if err != nil {
		return nil, err
	}

	return &task, nil
}


func (r *TaskRepository) UpdateTask(id int, updates map[string]interface{}) (*entities.Task, error) {
	var task entities.Task

	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}

	if len(updates) == 0 {
		return &task, nil
	}

	if err := r.db.Model(&task).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}

	return &task, nil
}
func (r *TaskRepository) DeleteTask(id int) error {
	err := r.db.
		Delete(&entities.Task{}, id).
		Error

	return err
}