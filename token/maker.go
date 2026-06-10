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
}
