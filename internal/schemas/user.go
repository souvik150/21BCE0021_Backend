package schemas

import "github.com/google/uuid"

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type RegisterUserInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserSignupResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string `json:"email"`
}

type GetCurrentUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}
