package util

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// ValidateArticleID validates that the given articleID is a valid UUID string,
// preventing path traversal attacks via crafted IDs like "../../etc".
func ValidateArticleID(articleID string) (uuid.UUID, error) {
	id, err := uuid.Parse(articleID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid article ID format: %w", err)
	}
	return id, nil
}

// SafeJoinPath joins path components under baseDir and verifies the result
// does not escape baseDir via ".." traversal or symlinks.
// Returns the cleaned absolute path if safe, or an error if the path escapes.
func SafeJoinPath(baseDir string, components ...string) (string, error) {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve base directory: %w", err)
	}

	joined := filepath.Join(append([]string{absBase}, components...)...)
	cleaned := filepath.Clean(joined)

	// Ensure the resulting path is still within baseDir
	if !strings.HasPrefix(cleaned, absBase+string(filepath.Separator)) && cleaned != absBase {
		return "", fmt.Errorf("path escapes base directory: %s", cleaned)
	}

	return cleaned, nil
}
