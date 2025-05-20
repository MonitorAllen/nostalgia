package token

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

// Payload contains the payload data of the token
type Payload struct {
	ID       uuid.UUID `json:"id"`
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	IssuedAt time.Time `json:"issued_at"`
	ExpireAt time.Time `json:"expire_at"`
}

func NewPayload(userID uuid.UUID, username string, role string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:       tokenID,
		UserID:   userID,
		Username: username,
		Role:     role,
		IssuedAt: time.Now(),
		ExpireAt: time.Now().Add(duration),
	}

	return payload, nil
}

// Valid checks if the token payload is valid
func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpireAt) {
		return ErrExpiredToken
	}
	return nil
}

type AdminPayload struct {
	ID       uuid.UUID `json:"id"`
	AdminID  int64     `json:"admin_id"`
	Username string    `json:"username"`
	RoleID   int64     `json:"role_id"`
	IssuedAt time.Time `json:"issued_at"`
	ExpireAt time.Time `json:"expire_at"`
}

func NewAdminPayload(adminID int64, username string, role_id int64, duration time.Duration) (*AdminPayload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &AdminPayload{
		ID:       tokenID,
		AdminID:  adminID,
		Username: username,
		RoleID:   role_id,
		IssuedAt: time.Now(),
		ExpireAt: time.Now().Add(duration),
	}

	return payload, nil
}

func (adminPayload *AdminPayload) Valid() error {
	if time.Now().After(adminPayload.ExpireAt) {
		return ErrExpiredToken
	}
	return nil
}
