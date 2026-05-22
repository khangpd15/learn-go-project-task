package entities

import "time"

type Project struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	OwnerID     int       `json:"owner_id" gorm:"not null;index"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewProject(name string, description string, ownerID int) Project {
	return Project{
		Name: name,
		Description: description,
		OwnerID: ownerID,
		CreatedAt: time.Now(),
	}
}