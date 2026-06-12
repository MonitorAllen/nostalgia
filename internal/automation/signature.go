package automation

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

var (
	ErrMissingSignatureField = errors.New("missing automation signature field")
	ErrInvalidKeyID          = errors.New("invalid automation key id")
	ErrInvalidTimestamp      = errors.New("invalid automation timestamp")
	ErrExpiredTimestamp      = errors.New("expired automation timestamp")
	ErrInvalidSignature      = errors.New("invalid automation signature")
)

type SignatureInput struct {
	Method         string
	Path           string
	Timestamp      string
	IdempotencyKey string
	Body           []byte
	Now            time.Time
	TTL            time.Duration
	KeyID          string
	ExpectedKeyID  string
	Secret         string
	Signature      string
}

func SHA256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func SignatureBaseString(method, path, timestamp, idempotencyKey, bodyHash string) string {
	return strings.Join([]string{method, path, timestamp, idempotencyKey, bodyHash}, "\n")
}

func Sign(secret, base string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(base))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifySignature(input SignatureInput) error {
	if input.Method == "" ||
		input.Path == "" ||
		input.Timestamp == "" ||
		input.IdempotencyKey == "" ||
		input.KeyID == "" ||
		input.ExpectedKeyID == "" ||
		input.Secret == "" ||
		input.Signature == "" {
		return ErrMissingSignatureField
	}

	if input.KeyID != input.ExpectedKeyID {
		return ErrInvalidKeyID
	}

	signedAt, err := time.Parse(time.RFC3339, input.Timestamp)
	if err != nil {
		return ErrInvalidTimestamp
	}

	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	ttl := input.TTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	if now.Sub(signedAt) > ttl || signedAt.Sub(now) > ttl {
		return ErrExpiredTimestamp
	}

	if !strings.HasPrefix(input.Signature, "v1=") {
		return ErrInvalidSignature
	}

	bodyHash := SHA256Hex(input.Body)
	base := SignatureBaseString(input.Method, input.Path, input.Timestamp, input.IdempotencyKey, bodyHash)
	expected := "v1=" + Sign(input.Secret, base)
	if !hmac.Equal([]byte(expected), []byte(input.Signature)) {
		return ErrInvalidSignature
	}

	return nil
}
