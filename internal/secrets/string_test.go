package secrets

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptStringRoundTripWithoutPlaintextLeak(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	aad := "nostalgia:ai-polish-api-key"
	plaintext := "sk-live-secret"

	ciphertext, err := EncryptString(plaintext, key, aad)
	require.NoError(t, err)
	require.NotEmpty(t, ciphertext)
	require.True(t, strings.HasPrefix(ciphertext, "v1:"))
	require.NotContains(t, ciphertext, plaintext)

	decrypted, err := DecryptString(ciphertext, key, aad)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

func TestDecryptStringRejectsWrongAAD(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	ciphertext, err := EncryptString("sk-live-secret", key, "nostalgia:ai-polish-api-key")
	require.NoError(t, err)

	_, err = DecryptString(ciphertext, key, "nostalgia:other-secret")
	require.Error(t, err)
}

func TestEncryptStringKeepsEmptySecretEmpty(t *testing.T) {
	ciphertext, err := EncryptString("", "0123456789abcdef0123456789abcdef", "aad")
	require.NoError(t, err)
	require.Empty(t, ciphertext)

	plaintext, err := DecryptString("", "0123456789abcdef0123456789abcdef", "aad")
	require.NoError(t, err)
	require.Empty(t, plaintext)
}
