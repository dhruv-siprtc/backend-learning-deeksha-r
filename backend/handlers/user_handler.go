package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"go-backend-learning/backend/config"
	"go-backend-learning/backend/messaging"
	"go-backend-learning/backend/models"
	"go-backend-learning/backend/request"
	"go-backend-learning/backend/response"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var validate = validator.New()

// CreateUser handles POST /users - Create a new user
func CreateUser(c echo.Context) error {
	var req request.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validate request
	if err := validate.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Check if user with this email already exists (excluding soft-deleted users)
	var existingUser models.User
	result := config.DB.Unscoped().Where("email = ? AND deleted_at IS NULL", req.Email).First(&existingUser)
	if result.Error == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Email already exists"})
	}

	// Create user
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	// Hash the password
	if err := user.HashPassword(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	if err := config.DB.Create(&user).Error; err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") ||
			strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Email already exists"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user", "details": err.Error()})
	}

	// Publish USER_CREATED event
	if messaging.Instance != nil {
		if err := messaging.PublishUserEvent(
			messaging.Instance,
			"user.created",
			"USER_CREATED",
			int(user.ID),
			user.Name,
			user.Email,
		); err != nil {
			// Log the error but don't fail the request since user creation was successful
			c.Logger().Errorf("Failed to publish USER_CREATED event: %v", err)
		}
	}

	return c.JSON(http.StatusCreated, response.ToUserResponse(user))
}

// GetUsers handles GET /users - Get all active users (exclude soft-deleted)
func GetUsers(c echo.Context) error {
	var users []models.User

	// GORM automatically excludes soft-deleted records
	if err := config.DB.Find(&users).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch users"})
	}

	// Convert to response format (exclude passwords)
	responses := make([]response.UserResponse, len(users))
	for i, user := range users {
		responses[i] = response.ToUserResponse(user)
	}

	return c.JSON(http.StatusOK, responses)
}

// GetUserByID handles GET /users/:id - Get user by ID (404 if soft-deleted)
func GetUserByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	var user models.User
	// GORM automatically excludes soft-deleted records
	if err := config.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch user"})
	}

	return c.JSON(http.StatusOK, response.ToUserResponse(user))
}

// UpdateUser handles PUT /users/:id - Update user (do not update soft-deleted users)
func UpdateUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	var req request.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validate request
	if err := validate.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Check if user exists and is not soft-deleted
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found or has been deleted"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch user"})
	}

	// If email is being updated, check if another active user already has this email
	if req.Email != nil {
		var existingUser models.User
		result := config.DB.Unscoped().Where("email = ? AND id != ? AND deleted_at IS NULL", *req.Email, id).First(&existingUser)
		if result.Error == nil {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Email already exists"})
		}
	}

	// Update only provided fields
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Password != nil {
		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
		}
		updates["password"] = string(hashedPassword)
	}

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") ||
			strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Email already exists"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user", "details": err.Error()})
	}

	// Fetch updated user
	config.DB.First(&user, id)

	// Publish USER_UPDATED event
	if messaging.Instance != nil {
		if err := messaging.PublishUserEvent(
			messaging.Instance,
			"user.updated",
			"USER_UPDATED",
			int(user.ID),
			user.Name,
			user.Email,
		); err != nil {
			// Log the error but don't fail the request since user update was successful
			c.Logger().Errorf("Failed to publish USER_UPDATED event: %v", err)
		}
	}

	return c.JSON(http.StatusOK, response.ToUserResponse(user))
}

// DeleteUser handles DELETE /users/:id - Soft delete user (set deleted_at)
func DeleteUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	// Check if user exists and is not already soft-deleted
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch user"})
	}

	// Soft delete the user (GORM sets deleted_at automatically)
	if err := config.DB.Delete(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
