package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const encryptedStringVersion = "v1"

var (
	ErrMissingKey        = errors.New("missing encryption key")
	ErrInvalidCiphertext = errors.New("invalid encrypted string")
)

func EncryptString(plaintext string, keyMaterial string, aad string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	aead, err := newAEAD(keyMaterial)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := aead.Seal(nil, nonce, []byte(plaintext), []byte(aad))
	encodedNonce := base64.RawStdEncoding.EncodeToString(nonce)
	encodedCiphertext := base64.RawStdEncoding.EncodeToString(ciphertext)
	return encryptedStringVersion + ":" + encodedNonce + ":" + encodedCiphertext, nil
}

func DecryptString(encrypted string, keyMaterial string, aad string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	parts := strings.Split(encrypted, ":")
	if len(parts) != 3 || parts[0] != encryptedStringVersion {
		return "", ErrInvalidCiphertext
	}

	nonce, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("%w: nonce", ErrInvalidCiphertext)
	}
	ciphertext, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return "", fmt.Errorf("%w: ciphertext", ErrInvalidCiphertext)
	}

	aead, err := newAEAD(keyMaterial)
	if err != nil {
		return "", err
	}
	if len(nonce) != aead.NonceSize() {
		return "", fmt.Errorf("%w: nonce size", ErrInvalidCiphertext)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, []byte(aad))
	if err != nil {
		return "", fmt.Errorf("%w: decrypt", ErrInvalidCiphertext)
	}
	return string(plaintext), nil
}

func newAEAD(keyMaterial string) (cipher.AEAD, error) {
	if strings.TrimSpace(keyMaterial) == "" {
		return nil, ErrMissingKey
	}

	key := sha256.Sum256([]byte("nostalgia-secrets:" + keyMaterial))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
