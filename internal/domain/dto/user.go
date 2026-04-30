package dto

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Email string `json:"email" validate:"required,email"`
}
