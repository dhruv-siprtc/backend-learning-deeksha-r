package response

import (
	"time"

	"go-backend-learning/backend/models"
)

type UserResponse struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ToUserResponse converts User model to UserResponse (excludes password)
func ToUserResponse(user models.User) UserResponse {
	var deletedAt *time.Time
	if user.DeletedAt.Valid {
		deletedAt = &user.DeletedAt.Time
	}

	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: deletedAt,
	}
}
