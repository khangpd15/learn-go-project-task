package services

import (
	"errors"
	"task_api/internal/entities"
	"task_api/internal/repositories"
	"task_api/internal/utils"
	"task_api/internal/validation"
)

var (
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrEmailAlreadyExist = errors.New("email already exists")
)

type UserService struct {
	userRepository repositories.UserRepositoryInterface
}

func NewUserService(userRepository repositories.UserRepositoryInterface) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (s *UserService) GetUserByID(id int) (*entities.User, error) {
	if !validation.IsValidIdUser(id) {
		return nil, ErrInvalidUserID
	}
	return s.userRepository.GetUserByID(id)
}

func (s *UserService) GetAllUsers() ([]entities.User, error) {
	return s.userRepository.GetAllUsers()
}

func (s *UserService) GetUserByFullName(fullName string) (*entities.User, error) {
	return s.userRepository.GetUserByFullName(fullName)
}

func (s *UserService) GetUserByEmail(email string) (*entities.User, error) {
	if !validation.IsValidEmail(email) {
		return nil, ErrInvalidEmail
	}
	return s.userRepository.GetUserByEmail(email)
}

func (s *UserService) CreateUser(user entities.User) (entities.User, error) {
	if !validation.IsValidPassword(user.PasswordHash) {
		return entities.User{}, ErrInvalidPassword
	}
	if !validation.IsValidEmail(user.Email) {
		return entities.User{}, ErrInvalidEmail
	}

	exists, err := s.userRepository.ExistsByEmail(user.Email)
	if err != nil {
		return entities.User{}, err
	}
	if exists {
		return entities.User{}, ErrEmailAlreadyExist
	}

	hashedPassword, err := utils.HashPassword(user.PasswordHash)
	if err != nil {
		return entities.User{}, err
	}

	user.PasswordHash = hashedPassword
	return s.userRepository.CreateUser(user)
}

func (s *UserService) UpdateUser(id int, updatedUser entities.User) (*entities.User, error) {
	if !validation.IsValidIdUser(id) {
		return nil, ErrInvalidUserID
	}

	if updatedUser.Email != "" {
		if !validation.IsValidEmail(updatedUser.Email) {
			return nil, ErrInvalidEmail
		}

		exists, err := s.userRepository.ExistsByEmailAndIDNot(updatedUser.Email, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrEmailAlreadyExist
		}
	}

	if updatedUser.PasswordHash != "" {
		if !validation.IsValidPassword(updatedUser.PasswordHash) {
			return nil, ErrInvalidPassword
		}

		hashedPassword, err := utils.HashPassword(updatedUser.PasswordHash)
		if err != nil {
			return nil, err
		}

		updatedUser.PasswordHash = hashedPassword
	}

	return s.userRepository.UpdateUser(id, updatedUser)
}

func (s *UserService) DeleteUser(id int) error {
	if !validation.IsValidIdUser(id) {
		return ErrInvalidUserID
	}
	return s.userRepository.DeleteUser(id)
}