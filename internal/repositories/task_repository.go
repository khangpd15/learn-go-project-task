package repositories
import (
	"task/api/internal/entities"
	"task/api/internal/data"
)
type TaskRepository struct {
}
func NewTaskRepository() *TaskRepository {
	return &TaskRepository{}
}
func (r *TaskRepository) GetAllTasks() []entities.Task {
	return data.Tasks
}
func (r *TaskRepository) GetTaskById(id int) *entities.Task{
	for _, task := range data.Tasks {
		if task.ID == id {	
			return &task
		}
	}
	return nil
}
func (r *TaskRepository) CreateTask(task entities.Task) entities.Task {
	task.ID = len(data.Tasks) + 1
	data.Tasks = append(data.Tasks, task)
	return task
}
func (r *TaskRepository) UpdateTask(id int, updatedTask entities.Task) *entities.Task {
	for i, task := range data.Tasks {	
		if task.ID == id {
			data.Tasks[i].Title = updatedTask.Title
			data.Tasks[i].Description = updatedTask.Description
			data.Tasks[i].Status = updatedTask.Status
			data.Tasks[i].Assignee = updatedTask.Assignee
			return &data.Tasks[i]
		}		
	}
	return nil
}
func (r *TaskRepository) DeleteTask(id int) bool {
	for i, task := range data.Tasks {
		if task.ID == id {					
			data.Tasks = append(data.Tasks[:i], data.Tasks[i+1:]...)
			return true
		}
	}
	return false
}