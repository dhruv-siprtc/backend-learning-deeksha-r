package manager

import (
	"log"
)

type UserManager struct{}

func (m *UserManager) VerifyUser(userID uint) {
	// Logic to verify user
	log.Printf("Verified user with ID: %d", userID)
}
