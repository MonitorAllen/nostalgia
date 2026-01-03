package cache

const (
	UserContributionsKey = "cache:user:contributions"
)

func GetUserContributionsKey() string {
	return UserContributionsKey
}
