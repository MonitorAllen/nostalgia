package util

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromEnvironmentWhenDotEnvIsMissing(t *testing.T) {
	configPath := t.TempDir() + string(os.PathSeparator)

	setConfigEnv(t, map[string]string{
		"ENVIRONMENT":              "production",
		"ALLOWED_ORIGINS":          "https://example.com,https://www.example.com",
		"DB_DRIVER":                "postgres",
		"DB_USER":                  "nostalgia",
		"DB_PASSWORD":              "secret",
		"DB_SOURCE":                "postgresql://nostalgia:secret@postgres:5432/nostalgia?sslmode=disable",
		"MIGRATION_URL":            "file://db/migration",
		"RESOURCE_PATH":            "./resources",
		"DOMAIN":                   "https://example.com",
		"REDIS_ADDRESS":            "redis:6379",
		"HTTP_SERVER_ADDRESS":      "0.0.0.0:8080",
		"GRPC_GATEWAY_ADDRESS":     "0.0.0.0:9091",
		"GRPC_SERVER_ADDRESS":      "0.0.0.0:9090",
		"TOKEN_SYMMETRIC_KEY":      "12345678901234567890123456789012",
		"SETUP_TOKEN":              "setup-token",
		"ACCESS_TOKEN_DURATION":    "15m",
		"REFRESH_TOKEN_DURATION":   "24h",
		"EMAIL_SENDER_NAME":        "Nostalgia",
		"EMAIL_SENDER_ADDRESS":     "noreply@example.com",
		"EMAIL_SENDER_PASSWORD":    "mail-secret",
		"UPLOAD_FILE_SIZE_LIMIT":   "5242880",
		"UPLOAD_FILE_ALLOWED_MIME": "image/jpeg,image/png",
		"HTTP_PROXY_ADDR":          "http://host.docker.internal:10808",
	})

	config, err := LoadConfig(configPath)
	require.NoError(t, err)

	require.Equal(t, "production", config.Environment)
	require.Equal(t, []string{"https://example.com", "https://www.example.com"}, config.AllowedOrigins)
	require.Equal(t, "postgresql://nostalgia:secret@postgres:5432/nostalgia?sslmode=disable", config.DBSource)
	require.Equal(t, "0.0.0.0:9091", config.GrpcGatewayAddress)
	require.Equal(t, 15*time.Minute, config.AccessTokenDuration)
	require.Equal(t, 24*time.Hour, config.RefreshTokenDuration)
	require.Equal(t, int64(5242880), config.UploadFileSizeLimit)
	require.Equal(t, []string{"image/jpeg", "image/png"}, config.UploadFileAllowedMime)
}

func TestLoadConfigEnvironmentOverridesDotEnvFile(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, ".env"), []byte("ENVIRONMENT=development\nHTTP_SERVER_ADDRESS=0.0.0.0:8080\n"), 0o600)
	require.NoError(t, err)

	t.Setenv("ENVIRONMENT", "production")

	config, err := LoadConfig(dir + string(os.PathSeparator))
	require.NoError(t, err)

	require.Equal(t, "production", config.Environment)
	require.Equal(t, "0.0.0.0:8080", config.HTTPServerAddress)
}

func TestConfigDoesNotExposeDefaultUserBootstrapEnv(t *testing.T) {
	removedKeys := []string{
		"DEFAULT_USER_ID",
		"DEFAULT_USERNAME",
		"DEFAULT_USER_PASSWORD",
		"DEFAULT_USER_FULLNAME",
		"DEFAULT_USER_EMAIL",
	}

	keys := configEnvKeys()
	for _, key := range removedKeys {
		require.NotContains(t, keys, key)
	}
}

func setConfigEnv(t *testing.T, values map[string]string) {
	t.Helper()

	for key, value := range values {
		t.Setenv(key, value)
	}
}
