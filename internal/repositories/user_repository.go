package repositories

import (
	"task/api/internal/data"
	"task/api/internal/entities"
)

type UserRepositoryInterface interface {
	GetUserByID(id int) *entities.User
}

type UserRepository struct{}

func NewUserRepository() UserRepositoryInterface {
	return &UserRepository{}
}

func (r *UserRepository) GetUserByID(id int) *entities.User {
	for i := range data.User {
		if data.User[i].ID == id {
			return &data.User[i]
		}
	}

	return nil
}