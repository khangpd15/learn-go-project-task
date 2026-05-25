package services

import "errors"

var (
	ErrForbidden          = errors.New("forbidden")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrTaskNotFound    = errors.New("task not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrProjectNotFound = errors.New("project not found")

	ErrInvalidTaskID         = errors.New("invalid task id")
	ErrInvalidUserID         = errors.New("invalid user id")
	ErrInvalidProjectID      = errors.New("invalid project id")
	ErrInvalidProjectOwnerId = errors.New("invalid project owner id")

	ErrInvalidStatus     = errors.New("invalid status")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrInvalidAssigneeID = errors.New("invalid assignee id")

	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrEmailAlreadyExist  = ErrEmailAlreadyExists
)
