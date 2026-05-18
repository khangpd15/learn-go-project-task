package services

import (
	"errors"
	"task_api/internal/entities"
	"task_api/internal/repositories"
)



type UserService struct {
	userRepository repositories.UserRepositoryInterface
}

func NewUserService(userRepository repositories.UserRepositoryInterface) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (s *UserService) GetUserByID(id int) *entities.User {
	return s.userRepository.GetUserByID(id)
}

func (s *UserService) GetAllUsers() []entities.User {
	return s.userRepository.GetAllUsers()
}

func (s *UserService) GetUserByUsername(username string) *entities.User {
	return s.userRepository.GetUserByUsername(username)
}

func (s *UserService) GetUserByEmail(email string) *entities.User {
	return s.userRepository.GetUserByEmail(email)
}

func (s *UserService) CreateUser(user entities.User) entities.User {
	return s.userRepository.CreateUser(user)
}

func (s *UserService) UpdateUser(id int, updatedUser entities.User) *entities.User {
	return s.userRepository.UpdateUser(id, updatedUser)
}

func (s *UserService) DeleteUser(id int) error {
	if !IsValidId(id) {
		return errors.New("invalid id")
	}
	if !s.userRepository.DeleteUser(id) {
		return errors.New("user not found")
	}
	return nil
}