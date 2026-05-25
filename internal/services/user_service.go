package services

import (
	requestDTO"task_api/internal/dto/request/user"
	"task_api/internal/entities"
	"task_api/internal/repositories"
	"task_api/internal/utils"
	"task_api/internal/validation"
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

func (s *UserService) CreateUser(req requestDTO.CreateUserRequest) (entities.User, error) {
	if !validation.IsValidPassword(req.Password) {
		return entities.User{}, ErrInvalidPassword
	}
	if !validation.IsValidEmail(req.Email) {
		return entities.User{}, ErrInvalidEmail
	}

	exists, err := s.userRepository.ExistsByEmail(req.Email)
	if err != nil {
		return entities.User{}, err
	}
	if exists {
		return entities.User{}, ErrEmailAlreadyExist
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return entities.User{}, err
	}

	user := entities.User{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}
	return s.userRepository.CreateUser(user)
}

func (s *UserService) UpdateUser(id int, req requestDTO.UpdateUserRequest) (*entities.User, error) {
	if !validation.IsValidIdUser(id) {
		return nil, ErrInvalidUserID
	}

	updatedUser := entities.User{}

	if req.FullName != nil {
		updatedUser.FullName = *req.FullName
	}

	if req.Email != nil && *req.Email != "" {
		if !validation.IsValidEmail(*req.Email) {
			return nil, ErrInvalidEmail
		}

		exists, err := s.userRepository.ExistsByEmailAndIDNot(*req.Email, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrEmailAlreadyExist
		}

		updatedUser.Email = *req.Email
	}

	if req.Password != nil && *req.Password != "" {
		if !validation.IsValidPassword(*req.Password) {
			return nil, ErrInvalidPassword
		}

		hashedPassword, err := utils.HashPassword(*req.Password)
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