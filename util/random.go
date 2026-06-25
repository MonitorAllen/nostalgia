package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	mathrand "math/rand"
	"strings"
	"time"
)

//goland:noinspection SpellCheckingInspection
const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	mathrand.Seed(time.Now().UnixNano())
}

// RandomInt generates a random integer between min and max.
// NOTE: Uses math/rand — suitable for tests and non-security purposes only.
func RandomInt(min, max int64) int64 {
	return min + mathrand.Int63n(max-min+1)
}

// RandomString generates a random string of length n.
// NOTE: Uses math/rand — suitable for tests and non-security purposes only.
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[mathrand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// SecureRandomString generates a cryptographically secure random hex string.
// The returned string has length 2*n (hex-encoded n random bytes).
// Use this for security-sensitive tokens (email verification, password resets, etc.).
func SecureRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto/rand read failed: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func RandUserID() uuid.UUID {
	UserID, err := uuid.NewRandom()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create userID")
	}
	return UserID
}

// RandomOwner generates a random owner name (test only)
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney generates a random amount of money (test only)
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency generates a random currency code (test only)
func RandomCurrency() string {
	currencies := []string{RMB, EUR, USD, CAD}
	n := len(currencies)
	return currencies[mathrand.Intn(n)]
}

// RandomEmail generates a random email (test only)
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}

func RandomTitle() string {
	return RandomString(6)
}

func RandomSummary() string {
	return RandomString(10)
}

func RandomContext() string {
	return RandomString(100)
}
