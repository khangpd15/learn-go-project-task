package services

import (
	"errors"
	"testing"

	requestauth "task_api/internal/dto/request/auth"
	"task_api/internal/entities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	getUserByEmailFn func(email string) (*entities.User, error)
	existsByEmailFn  func(email string) (bool, error)
	createUserFn     func(user entities.User) (entities.User, error)
}

func (m *mockUserRepository) GetUserByID(id int) (*entities.User, error) { return nil, nil }
func (m *mockUserRepository) GetAllUsers() ([]entities.User, error)      { return nil, nil }
func (m *mockUserRepository) GetUserByFullName(fullName string) (*entities.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetUserByEmail(email string) (*entities.User, error) {
	if m.getUserByEmailFn != nil {
		return m.getUserByEmailFn(email)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserRepository) CreateUser(user entities.User) (entities.User, error) {
	if m.createUserFn != nil {
		return m.createUserFn(user)
	}
	return entities.User{}, errors.New("not implemented")
}

func (m *mockUserRepository) UpdateUser(id int, updatedUser entities.User) (*entities.User, error) {
	return nil, nil
}
func (m *mockUserRepository) DeleteUser(id int) error { return nil }
func (m *mockUserRepository) ExistsByEmail(email string) (bool, error) {
	if m.existsByEmailFn != nil {
		return m.existsByEmailFn(email)
	}
	return false, errors.New("not implemented")
}
func (m *mockUserRepository) ExistsByEmailAndIDNot(email string, id int) (bool, error) {
	return false, nil
}

func TestAuthService_Login_Success(t *testing.T) {
	password := "Password@123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	repo := &mockUserRepository{
		getUserByEmailFn: func(email string) (*entities.User, error) {
			require.Equal(t, "user@example.com", email)
			return &entities.User{ID: 11, Email: email, PasswordHash: string(hashedPassword)}, nil
		},
	}

	service := NewAuthService(repo)

	token, err := service.Login(requestauth.LoginRequest{Email: "user@example.com", Password: password})

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password@123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	repo := &mockUserRepository{
		getUserByEmailFn: func(email string) (*entities.User, error) {
			return &entities.User{ID: 11, Email: email, PasswordHash: string(hashedPassword)}, nil
		},
	}

	service := NewAuthService(repo)

	token, err := service.Login(requestauth.LoginRequest{Email: "user@example.com", Password: "Wrong@123"})

	assert.Error(t, err)
	assert.Equal(t, "invalid email or password", err.Error())
	assert.Empty(t, token)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := &mockUserRepository{
		getUserByEmailFn: func(email string) (*entities.User, error) {
			return nil, errors.New("user not found")
		},
	}

	service := NewAuthService(repo)

	token, err := service.Login(requestauth.LoginRequest{Email: "missing@example.com", Password: "Password@123"})

	assert.Error(t, err)
	assert.Equal(t, "invalid email or password", err.Error())
	assert.Empty(t, token)
}

func TestAuthService_Register_Success(t *testing.T) {
	var createdUser entities.User

	repo := &mockUserRepository{
		existsByEmailFn: func(email string) (bool, error) {
			return false, nil
		},
		createUserFn: func(user entities.User) (entities.User, error) {
			createdUser = user
			user.ID = 22
			return user, nil
		},
	}

	service := NewAuthService(repo)
	req := requestauth.RegisterRequest{FullName: "Test User", Email: "test@example.com", Password: "Password@123"}

	err := service.Register(req)

	require.NoError(t, err)
	assert.Equal(t, req.FullName, createdUser.FullName)
	assert.Equal(t, req.Email, createdUser.Email)
	assert.NotEqual(t, req.Password, createdUser.PasswordHash)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(createdUser.PasswordHash), []byte(req.Password)))
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	repo := &mockUserRepository{
		existsByEmailFn: func(email string) (bool, error) { return true, nil },
	}

	service := NewAuthService(repo)
	err := service.Register(requestauth.RegisterRequest{FullName: "Test User", Email: "test@example.com", Password: "Password@123"})

	assert.Error(t, err)
	assert.Equal(t, "email already exists", err.Error())
}

func TestAuthService_Register_InvalidInputs(t *testing.T) {
	repo := &mockUserRepository{
		existsByEmailFn: func(email string) (bool, error) { return false, nil },
	}

	service := NewAuthService(repo)

	tests := []struct {
		name string
		req  requestauth.RegisterRequest
		want string
	}{
		{name: "missing field", req: requestauth.RegisterRequest{FullName: "", Email: "test@example.com", Password: "Password@123"}, want: "all fields are required"},
		{name: "invalid email", req: requestauth.RegisterRequest{FullName: "Test User", Email: "invalid-email", Password: "Password@123"}, want: "invalid email format"},
		{name: "invalid password", req: requestauth.RegisterRequest{FullName: "Test User", Email: "test@example.com", Password: "short"}, want: "password must be at least 8 characters and contain at least one letter and one number and one special character"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Register(tt.req)
			assert.Error(t, err)
			assert.Equal(t, tt.want, err.Error())
		})
	}
}
