package repositories

import (
	"task_api/internal/data"
	"task_api/internal/entities"
)

type UserRepositoryInterface interface {
	GetUserByID(id int) *entities.User
	GetAllUsers() []entities.User
	GetUserByUsername(username string) *entities.User
	GetUserByEmail(email string) *entities.User
	CreateUser(user entities.User) entities.User
	UpdateUser(id int, updatedUser entities.User) *entities.User
	DeleteUser(id int) bool
}

type UserRepository struct{}

func NewUserRepository() UserRepositoryInterface {
	return &UserRepository{}
}
func (r *UserRepository) GetAllUsers() []entities.User {
	return data.User
}
func (r *UserRepository) GetUserByUsername(username string) *entities.User {
	for i := range data.User {
		if data.User[i].Username == username {
			return &data.User[i]
		}	
	}
	return nil
}
func (r*UserRepository) GetUserByEmail(email string) *entities.User {
	for i := range data.User {
		if data.User[i].Email == email {
			return &data.User[i]
		}		
	}
	return nil
}

func (r *UserRepository) CreateUser(user entities.User) entities.User {
	user.ID = len(data.User) + 1
	data.User = append(data.User, user)
	return user
}	
func (r *UserRepository) UpdateUser(id int, updatedUser entities.User) *entities.User {
	for i := range data.User {	
		if data.User[i].ID == id {
			data.User[i].Role = updatedUser.Role
			data.User[i].Username = updatedUser.Username
			data.User[i].Password = updatedUser.Password
			data.User[i].Email = updatedUser.Email
			return &data.User[i]
		}	
	}
	return nil
}
func (r *UserRepository) DeleteUser(id int) bool {
	for i := range data.User {
		if data.User[i].ID == id {	
			data.User = append(data.User[:i], data.User[i+1:]...)
			return true
		}
	}
	return false
}

func (r *UserRepository) GetUserByID(id int) *entities.User {
	for i := range data.User {
		if data.User[i].ID == id {
			return &data.User[i]
		}
	}

	return nil
}