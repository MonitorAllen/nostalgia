package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateArticleID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid UUID", "550e8400-e29b-41d4-a716-446655440000", false},
		{"path traversal attempt", "../../etc", true},
		{"encoded traversal", "..%2F..%2Fetc", true},
		{"empty string", "", true},
		{"plain text", "not-a-uuid", true},
		{"partial UUID", "550e8400-e29b-41d4", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateArticleID(tc.input)
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSafeJoinPath(t *testing.T) {
	tests := []struct {
		name       string
		baseDir    string
		components []string
		wantError  bool
	}{
		{"normal path", "/tmp/resources", []string{"articles", "abc"}, false},
		{"traversal attempt", "/tmp/resources", []string{"..", "..", "etc", "passwd"}, true},
		{"traversal in component", "/tmp/resources", []string{"articles", "../../../etc"}, true},
		{"single dot resolves to base", "/tmp/resources", []string{"."}, false},
		{"empty component resolves to base", "/tmp/resources", []string{""}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := SafeJoinPath(tc.baseDir, tc.components...)
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Contains(t, result, tc.baseDir)
			}
		})
	}
}
