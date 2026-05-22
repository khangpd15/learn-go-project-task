package repositories

import(
	"task_api/internal/entities"
	"gorm.io/gorm"
)


type ProjectRepositoryInterface interface {
	ListProjectByOwner(ownerID int) ([]entities.Project, error)
	ListAllProjects() ([]entities.Project, error)
	GetProjectByID(projectID int) (entities.Project, error)
	UpdateProject(projectID int, name string, description string) error
	DeleteProject(project entities.Project) error
}
type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) ListProjectByOwner(ownerID int) ([]entities.Project, error) {
	var projects []entities.Project
	err := r.db.Where("owner_id = ?", ownerID).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) ListAllProjects() ([]entities.Project, error) {
	var projects []entities.Project
	err := r.db.Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) UpdateProject(projectID int, name string, description string) error {
	var project entities.Project
	err := r.db.First(&project, projectID).Error
	if err != nil {
		return err
	}
	project.Name = name
	project.Description = description
	return r.db.Save(&project).Error
}
func (r *ProjectRepository) GetProjectByID(projectID int) (entities.Project, error) {
	var project entities.Project
	err := r.db.First(&project, projectID).Error
	if err != nil {
		return entities.Project{}, err
	}
	return project, nil
}

func (r *ProjectRepository) DeleteProject(project entities.Project) error {
	return r.db.Delete(&project).Error
}