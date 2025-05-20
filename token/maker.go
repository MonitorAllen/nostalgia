package token

import (
	"github.com/google/uuid"
	"time"
)

// Maker is an interface for managing tokens
type Maker interface {
	// CreateToken creates a new user token for a specific username and duration
	CreateToken(UserID uuid.UUID, username string, role string, duration time.Duration) (string, *Payload, error)

	// VerifyToken checks if the user token is valid or not
	VerifyToken(token string) (*Payload, error)

	// CreateAdminToken creates a new admin token for a specific username and duration
	CreateAdminToken(AdminID int64, username string, roleId int64, duration time.Duration) (string, *AdminPayload, error)

	// VerifyAdminToken checks if the admin token is valid or not
	VerifyAdminToken(token string) (*AdminPayload, error)
}
