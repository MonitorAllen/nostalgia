package automation

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVerifySignatureAcceptsValidRequest(t *testing.T) {
	body := []byte(`{"title":"Go cache"}`)
	timestamp := "2026-06-12T10:30:00+08:00"
	bodyHash := SHA256Hex(body)
	base := SignatureBaseString("POST", "/api/automation/articles/drafts", timestamp, "daily-go-cache", bodyHash)
	signature := "v1=" + Sign("secret", base)

	err := VerifySignature(SignatureInput{
		Method:         "POST",
		Path:           "/api/automation/articles/drafts",
		Timestamp:      timestamp,
		IdempotencyKey: "daily-go-cache",
		Body:           body,
		Now:            time.Date(2026, 6, 12, 10, 32, 0, 0, time.FixedZone("CST", 8*60*60)),
		TTL:            5 * time.Minute,
		KeyID:          "codex-daily-writer",
		ExpectedKeyID:  "codex-daily-writer",
		Secret:         "secret",
		Signature:      signature,
	})
	require.NoError(t, err)
}

func TestVerifySignatureRejectsInvalidSignature(t *testing.T) {
	err := VerifySignature(SignatureInput{
		Method:         "POST",
		Path:           "/api/automation/articles/drafts",
		Timestamp:      "2026-06-12T10:30:00+08:00",
		IdempotencyKey: "daily-go-cache",
		Body:           []byte(`{"title":"Go cache"}`),
		Now:            time.Date(2026, 6, 12, 10, 31, 0, 0, time.FixedZone("CST", 8*60*60)),
		TTL:            5 * time.Minute,
		KeyID:          "codex-daily-writer",
		ExpectedKeyID:  "codex-daily-writer",
		Secret:         "secret",
		Signature:      "v1=bad",
	})
	require.ErrorIs(t, err, ErrInvalidSignature)
}

func TestVerifySignatureRejectsExpiredTimestamp(t *testing.T) {
	body := []byte(`{"title":"Go cache"}`)
	timestamp := "2026-06-12T10:20:00+08:00"
	base := SignatureBaseString("POST", "/api/automation/articles/drafts", timestamp, "daily-go-cache", SHA256Hex(body))

	err := VerifySignature(SignatureInput{
		Method:         "POST",
		Path:           "/api/automation/articles/drafts",
		Timestamp:      timestamp,
		IdempotencyKey: "daily-go-cache",
		Body:           body,
		Now:            time.Date(2026, 6, 12, 10, 30, 0, 0, time.FixedZone("CST", 8*60*60)),
		TTL:            5 * time.Minute,
		KeyID:          "codex-daily-writer",
		ExpectedKeyID:  "codex-daily-writer",
		Secret:         "secret",
		Signature:      "v1=" + Sign("secret", base),
	})
	require.ErrorIs(t, err, ErrExpiredTimestamp)
}

func TestVerifySignatureRejectsMissingRequiredFields(t *testing.T) {
	err := VerifySignature(SignatureInput{})
	require.ErrorIs(t, err, ErrMissingSignatureField)
}

func TestSignatureBaseStringIncludesCanonicalParts(t *testing.T) {
	base := SignatureBaseString("POST", "/api/automation/articles/drafts", "2026-06-12T10:30:00+08:00", "daily-go-cache", "abc123")

	require.Equal(t, strings.Join([]string{
		"POST",
		"/api/automation/articles/drafts",
		"2026-06-12T10:30:00+08:00",
		"daily-go-cache",
		"abc123",
	}, "\n"), base)
}
