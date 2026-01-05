package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-backend-learning/backend/config"
	"go-backend-learning/backend/models"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	// Init test DB
	config.InitTestDB()
	config.DB.AutoMigrate(&models.User{})

	e := echo.New()

	reqBody := `{
		"name": "Test User",
		"email": "testuser1@example.com",
		"password": "password123"
	}`

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	// Call handler directly
	err := CreateUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}
