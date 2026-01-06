package key

import "fmt"

const (
	AdminSessionKey = "cache:admin:session:%d"
	AdminMenuKey    = "cache:admin:menu:%d"
)

func GetAdminSessionKey(adminID int64) string {
	return fmt.Sprintf(AdminSessionKey, adminID)
}

func GetAdminMenuKey(roleID int64) string {
	return fmt.Sprintf(AdminMenuKey, roleID)
}
