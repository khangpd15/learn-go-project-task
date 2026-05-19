package mapper

import (
	requestDTO "task_api/internal/dto/request/user"
	responseDTO "task_api/internal/dto/response/user"
	"task_api/internal/entities"
)

func UserToEntity(req requestDTO.CreateUserRequest) entities.User {
	return entities.User{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: req.Password,
	}
}

func ToUpdateUserEntity(req requestDTO.UpdateUserRequest) entities.User {
	user := entities.User{}

	if req.FullName != nil {
		user.FullName = *req.FullName
	}

	if req.Email != nil {
		user.Email = *req.Email
	}

	if req.Password != nil {
		user.PasswordHash = *req.Password
	}

	return user
}

func UserToResponse(user entities.User) responseDTO.UserResponse {
	return responseDTO.UserResponse{
		ID:       user.ID,
		FullName: user.FullName,
		Email:    user.Email,
	}
}

func UsersToResponses(users []entities.User) []responseDTO.UserResponse {
	responses := make([]responseDTO.UserResponse, 0, len(users))

	for _, user := range users {
		responses = append(responses, UserToResponse(user))
	}

	return responses
}