package services

import (
	"errors"
	"fmt"
	"strings"
	"task_api/internal/dto/request/auth"
	"task_api/internal/entities"
	"task_api/internal/repositories"
	"task_api/internal/utils"
	"task_api/internal/validation"

	"github.com/jackc/pgx/v5/pgconn"
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
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return "", ErrInvalidInput
	}

	user, err := as.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := utils.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", fmt.Errorf("generate access token: %w", err)
	}

	return token, nil
}
func (as *AuthService) Register(req auth.RegisterRequest) error {

	if strings.TrimSpace(req.FullName) == "" || strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return ErrInvalidInput
	}

	if !validation.IsValidEmail(req.Email) {
		return ErrInvalidEmail
	}
	if !validation.IsValidPassword(req.Password) {
		return ErrInvalidPassword
	}
	exists, err := as.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return err
	}

	if exists {
		return ErrEmailAlreadyExists
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	newUser := entities.NewUser(
		0,
		strings.TrimSpace(req.FullName),
		string(hashedPassword),
		strings.TrimSpace(req.Email),
	)
	_, err = as.userRepo.CreateUser(newUser)
	if err != nil {
		if isUniqueEmailViolation(err) {
			return ErrEmailAlreadyExists
		}
		return err
	}
	return nil
}

func isUniqueEmailViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}

	return false
}
