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
	require.Equal(t, 0, config.RedisCacheDB)
	require.Equal(t, 1, config.RedisQueueDB)
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

func TestLoadConfigRedisDatabaseOverrides(t *testing.T) {
	configPath := t.TempDir() + string(os.PathSeparator)

	setConfigEnv(t, map[string]string{
		"REDIS_CACHE_DB": "2",
		"REDIS_QUEUE_DB": "3",
	})

	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	require.Equal(t, 2, config.RedisCacheDB)
	require.Equal(t, 3, config.RedisQueueDB)
}

func TestConfigExposesRedisDatabaseEnvKeys(t *testing.T) {
	keys := configEnvKeys()
	require.Contains(t, keys, "REDIS_CACHE_DB")
	require.Contains(t, keys, "REDIS_QUEUE_DB")
}

func TestLoadConfigAutomationOverrides(t *testing.T) {
	configPath := t.TempDir() + string(os.PathSeparator)

	setConfigEnv(t, map[string]string{
		"AUTOMATION_HMAC_KEY_ID":       "codex-daily-writer",
		"AUTOMATION_HMAC_SECRET":       "secret",
		"AUTOMATION_SIGNATURE_TTL":     "10m",
		"AUTOMATION_DAILY_DRAFT_LIMIT": "3",
		"AUTOMATION_NOTIFY_EMAIL":      "owner@example.com",
	})

	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	require.Equal(t, "codex-daily-writer", config.AutomationHMACKeyID)
	require.Equal(t, "secret", config.AutomationHMACSecret)
	require.Equal(t, 10*time.Minute, config.AutomationSignatureTTL)
	require.Equal(t, int64(3), config.AutomationDailyDraftLimit)
	require.Equal(t, "owner@example.com", config.AutomationNotifyEmail)
}

func TestLoadConfigAutomationDefaults(t *testing.T) {
	configPath := t.TempDir() + string(os.PathSeparator)

	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	require.Equal(t, 5*time.Minute, config.AutomationSignatureTTL)
	require.Equal(t, int64(1), config.AutomationDailyDraftLimit)
}

func TestConfigExposesAutomationEnvKeys(t *testing.T) {
	keys := configEnvKeys()
	require.Contains(t, keys, "AUTOMATION_HMAC_KEY_ID")
	require.Contains(t, keys, "AUTOMATION_HMAC_SECRET")
	require.Contains(t, keys, "AUTOMATION_SIGNATURE_TTL")
	require.Contains(t, keys, "AUTOMATION_DAILY_DRAFT_LIMIT")
	require.Contains(t, keys, "AUTOMATION_NOTIFY_EMAIL")
}

func TestLoadConfigAIPolishOverrides(t *testing.T) {
	configPath := t.TempDir() + string(os.PathSeparator)

	setConfigEnv(t, map[string]string{
		"AI_POLISH_PROVIDER":          "openai_compatible",
		"AI_POLISH_BASE_URL":          "https://ai.example.com/v1",
		"AI_POLISH_API_KEY":           "runtime-secret",
		"AI_POLISH_MODEL":             "writer-model",
		"AI_POLISH_TIMEOUT":           "45s",
		"AI_POLISH_MAX_INPUT_CHARS":   "7000",
		"AI_POLISH_MAX_CONTEXT_CHARS": "5000",
		"AI_POLISH_MAX_SUGGESTIONS":   "2",
	})

	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	require.Equal(t, "openai_compatible", config.AIPolishProvider)
	require.Equal(t, "https://ai.example.com/v1", config.AIPolishBaseURL)
	require.Equal(t, "runtime-secret", config.AIPolishAPIKey)
	require.Equal(t, "writer-model", config.AIPolishModel)
	require.Equal(t, 45*time.Second, config.AIPolishTimeout)
	require.Equal(t, 7000, config.AIPolishMaxInputChars)
	require.Equal(t, 5000, config.AIPolishMaxContextChars)
	require.Equal(t, 2, config.AIPolishMaxSuggestions)
}

func TestLoadConfigAIPolishDefaults(t *testing.T) {
	configPath := t.TempDir() + string(os.PathSeparator)

	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	require.Equal(t, "openai_compatible", config.AIPolishProvider)
	require.Equal(t, 30*time.Second, config.AIPolishTimeout)
	require.Equal(t, 6000, config.AIPolishMaxInputChars)
	require.Equal(t, 4000, config.AIPolishMaxContextChars)
	require.Equal(t, 3, config.AIPolishMaxSuggestions)
}

func TestConfigExposesAIPolishEnvKeys(t *testing.T) {
	keys := configEnvKeys()
	require.Contains(t, keys, "AI_POLISH_PROVIDER")
	require.Contains(t, keys, "AI_POLISH_BASE_URL")
	require.Contains(t, keys, "AI_POLISH_API_KEY")
	require.Contains(t, keys, "AI_POLISH_MODEL")
	require.Contains(t, keys, "AI_POLISH_TIMEOUT")
	require.Contains(t, keys, "AI_POLISH_MAX_INPUT_CHARS")
	require.Contains(t, keys, "AI_POLISH_MAX_CONTEXT_CHARS")
	require.Contains(t, keys, "AI_POLISH_MAX_SUGGESTIONS")
}

func setConfigEnv(t *testing.T, values map[string]string) {
	t.Helper()

	for key, value := range values {
		t.Setenv(key, value)
	}
}
