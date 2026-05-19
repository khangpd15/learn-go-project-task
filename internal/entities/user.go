package entities

import "time"

type User struct {
	ID           int       `json:"id"`
	FullName     string    `json:"fullname"`
	PasswordHash string    `json:"-"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
}

func NewUser(id int, fullName string, passwordHash string, email string) User {
	return User{
		ID:           id,
		FullName:     fullName,
		PasswordHash: passwordHash,
		Email:        email,
		CreatedAt:    time.Now(),
	}
}
