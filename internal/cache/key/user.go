package key

import "fmt"

const (
	UserContributionsKey = "cache:user:contributions"
	UserDisabledPrefix   = "cache:user:disabled:"
)

func GetUserContributionsKey() string {
	return UserContributionsKey
}

// GetUserDisabledKey returns the cache key for a disabled user flag.
// Presence of this key means the user is disabled.
func GetUserDisabledKey(userID string) string {
	return fmt.Sprintf("%s%s", UserDisabledPrefix, userID)
}
