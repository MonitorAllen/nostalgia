package util

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"math/rand"
	"strings"
	"time"
)

//goland:noinspection SpellCheckingInspection
const alphabet = "abcdefghijklmnopqrstuvwsyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomInt generates a random integer between min an max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func RandUserID() uuid.UUID {
	UserID, err := uuid.NewRandom()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create userID")
	}
	return UserID
}

// RandomOwner generates a random owner name
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney generates a random amount of money
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency generates a random currency code
func RandomCurrency() string {
	currencies := []string{RMB, EUR, USD, CAD}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}

// RandomEmail generates a random email
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
