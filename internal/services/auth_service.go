package services

import (
	"errors"
	"fmt"
	"task_api/internal/dto/request/auth"
	"task_api/internal/entities"
	"task_api/internal/repositories"
	"task_api/internal/utils"
	"task_api/internal/validation"

	"golang.org/x/crypto/bcrypt"
)

type AuthServiceInterface interface {
	Login(req auth.LoginRequest) (string, error)
	Register(req auth.RegisterRequest) error
}

type AuthService struct {
	userRepo repositories.UserRepositoryInterface
}

func NewAuthService(userRepo repositories.UserRepositoryInterface) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (as *AuthService) Login(req auth.LoginRequest) (string, error) {
	user, err := as.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}
fmt.Println("email:", user.Email)
fmt.Println("hash:", user.PasswordHash)
fmt.Println("password req:", req.Password)

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	token, err := utils.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", errors.New("failed to generate access token")
	}

	return token, nil
}
func (as *AuthService) Register(req auth.RegisterRequest) error {
	_, err := as.userRepo.GetUserByEmail(req.Email)
	if err == nil {
		return errors.New("email already exists")
	}
	if !validation.IsValidEmail(req.Email) {
		return errors.New("invalid email format")
	}
	if !validation.IsValidPassword(req.Password) {
		return errors.New("password must be at least 8 characters and contain at least one letter and one number and one special character")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}
	if req.FullName == "" || req.Email == "" || req.Password == "" {
		return errors.New("all fields are required")
	}
	newUser := entities.NewUser(
		0,
		req.FullName,
		string(hashedPassword),
		req.Email,
	)
    _, err = as.userRepo.CreateUser(newUser)
    if err != nil {
        return errors.New("failed to create user")
    }
    return nil
}