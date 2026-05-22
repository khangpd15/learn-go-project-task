package repositories

import (
	"task_api/internal/entities"
	"errors"

	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	GetUserByID(id int) (*entities.User, error)
	GetAllUsers() ([]entities.User, error)
	GetUserByFullName(fullName string) (*entities.User, error)
	GetUserByEmail(email string) (*entities.User, error)
	CreateUser(user entities.User) (entities.User, error)
	UpdateUser(id int, updatedUser entities.User) (*entities.User, error)
	DeleteUser(id int) error
	ExistsByEmail(email string) (bool, error)
	ExistsByEmailAndIDNot(email string, id int) (bool, error)
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetAllUsers() ([]entities.User, error) {
	var users []entities.User

	err := r.db.Order("id asc").Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) GetUserByFullName(fullName string) (*entities.User, error) {
	var user entities.User

	err := r.db.Where("full_name = ?", fullName).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*entities.User, error) {
	var user entities.User

	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) CreateUser(user entities.User) (entities.User, error) {
	err := r.db.Create(&user).Error
	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *UserRepository) UpdateUser(id int, updatedUser entities.User) (*entities.User, error) {
	var user entities.User

	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}

	if updatedUser.FullName != "" {
		user.FullName = updatedUser.FullName
	}

	if updatedUser.Email != "" {
		user.Email = updatedUser.Email
	}

	if updatedUser.PasswordHash != "" {
		user.PasswordHash = updatedUser.PasswordHash
	}

	err = r.db.Save(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) DeleteUser(id int) error {
	var user entities.User

	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return err
	}

	err = r.db.Delete(&user).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) GetUserByID(id int) (*entities.User, error) {
	var user entities.User

	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}
func (r *UserRepository) ExistsByEmailAndIDNot(email string, id int) (bool, error) {
	var count int64

	err := r.db.Model(&entities.User{}).
		Where("email = ? AND id <> ?", email, id).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&entities.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}