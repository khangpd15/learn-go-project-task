package services
import
(
	"errors"
	"task/api/internal/entities"
	"task/api/internal/repositories"
)
type TaskService struct {
	repo *repositories.TaskRepository
}
func NewTaskService(repo *repositories.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func IsValidStatus(status string) bool {
	validStatus := map[string]bool{
		"TODO": true,
		"IN_PROGRESS": true,
		"DONE": true,
	}
	return validStatus[status]
}
func IsValidId(id int) bool {

		return id > 0
}

 func (s *TaskService) GetAllTasks() []entities.Task{
	return s.repo.GetAllTasks()
 }

 func (s *TaskService) GetTaskById(id int) *entities.Task{
	return s.repo.GetTaskById(id)
 }
 
 func (s *TaskService) CreateTask(task entities.Task) (entities.Task, error){
	if !IsValidStatus(task.Status) {
		return entities.Task{}, errors.New("invalid status")
	}
	return s.repo.CreateTask(task), nil
}
func (s *TaskService) UpdateTask(id int, updateTask entities.Task) (*entities.Task, error){
	if !IsValidId(id) {
		return nil, errors.New("invalid id")
	}
	if !IsValidStatus(updateTask.Status) {
		return nil, errors.New("invalid status")
	}
	return s.repo.UpdateTask(id, updateTask), nil
}
func (s *TaskService) DeleteTask(id int) error{
	if !IsValidId(id) {
		return errors.New("invalid id")
	}	
	if !s.repo.DeleteTask(id) {
		return errors.New("task not found")
	}
	return nil
}
