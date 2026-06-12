package util

import (
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment               string        `mapstructure:"ENVIRONMENT"`
	AllowedOrigins            []string      `mapstructure:"ALLOWED_ORIGINS"`
	DBDriver                  string        `mapstructure:"DB_DRIVER"`
	DBUser                    string        `mapstructure:"DB_USER"`
	DBPassword                string        `mapstructure:"DB_PASSWORD"`
	DBSource                  string        `mapstructure:"DB_SOURCE"`
	MigrationURL              string        `mapstructure:"MIGRATION_URL"`
	ResourcePath              string        `mapstructure:"RESOURCE_PATH"`
	Domain                    string        `mapstructure:"DOMAIN"`
	RedisAddress              string        `mapstructure:"REDIS_ADDRESS"`
	RedisCacheDB              int           `mapstructure:"REDIS_CACHE_DB"`
	RedisQueueDB              int           `mapstructure:"REDIS_QUEUE_DB"`
	AutomationHMACKeyID       string        `mapstructure:"AUTOMATION_HMAC_KEY_ID"`
	AutomationHMACSecret      string        `mapstructure:"AUTOMATION_HMAC_SECRET"`
	AutomationSignatureTTL    time.Duration `mapstructure:"AUTOMATION_SIGNATURE_TTL"`
	AutomationDailyDraftLimit int64         `mapstructure:"AUTOMATION_DAILY_DRAFT_LIMIT"`
	AutomationNotifyEmail     string        `mapstructure:"AUTOMATION_NOTIFY_EMAIL"`
	AIPolishProvider          string        `mapstructure:"AI_POLISH_PROVIDER"`
	AIPolishBaseURL           string        `mapstructure:"AI_POLISH_BASE_URL"`
	AIPolishAPIKey            string        `mapstructure:"AI_POLISH_API_KEY"`
	AIPolishModel             string        `mapstructure:"AI_POLISH_MODEL"`
	AIPolishTimeout           time.Duration `mapstructure:"AI_POLISH_TIMEOUT"`
	AIPolishMaxInputChars     int           `mapstructure:"AI_POLISH_MAX_INPUT_CHARS"`
	AIPolishMaxContextChars   int           `mapstructure:"AI_POLISH_MAX_CONTEXT_CHARS"`
	AIPolishMaxSuggestions    int           `mapstructure:"AI_POLISH_MAX_SUGGESTIONS"`
	HTTPServerAddress         string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	GrpcGatewayAddress        string        `mapstructure:"GRPC_GATEWAY_ADDRESS"`
	GRPCServerAddress         string        `mapstructure:"GRPC_SERVER_ADDRESS"`
	TokenSymmetricKey         string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	SetupToken                string        `mapstructure:"SETUP_TOKEN"`
	AccessTokenDuration       time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration      time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	EmailSenderName           string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress        string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword       string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
	UploadFileSizeLimit       int64         `mapstructure:"UPLOAD_FILE_SIZE_LIMIT"`
	UploadFileAllowedMime     []string      `mapstructure:"UPLOAD_FILE_ALLOWED_MIME"`
	HTTPProxyAddr             string        `mapstructure:"HTTP_PROXY_ADDR"`
}

func LoadConfig(path string) (config Config, err error) {
	configReader := viper.New()
	configReader.SetConfigFile(filepath.Join(path, ".env"))
	configReader.AutomaticEnv()
	configReader.SetDefault("REDIS_CACHE_DB", 0)
	configReader.SetDefault("REDIS_QUEUE_DB", 1)
	configReader.SetDefault("AUTOMATION_SIGNATURE_TTL", 5*time.Minute)
	configReader.SetDefault("AUTOMATION_DAILY_DRAFT_LIMIT", 1)
	configReader.SetDefault("AI_POLISH_PROVIDER", "openai_compatible")
	configReader.SetDefault("AI_POLISH_TIMEOUT", 30*time.Second)
	configReader.SetDefault("AI_POLISH_MAX_INPUT_CHARS", 6000)
	configReader.SetDefault("AI_POLISH_MAX_CONTEXT_CHARS", 4000)
	configReader.SetDefault("AI_POLISH_MAX_SUGGESTIONS", 3)

	for _, key := range configEnvKeys() {
		if bindErr := configReader.BindEnv(key); bindErr != nil {
			return config, bindErr
		}
	}

	err = configReader.ReadInConfig()
	if err != nil && !os.IsNotExist(err) {
		return
	}

	err = configReader.Unmarshal(&config)
	return
}

func configEnvKeys() []string {
	configType := reflect.TypeOf(Config{})
	keys := make([]string, 0, configType.NumField())

	for i := 0; i < configType.NumField(); i++ {
		if key := configType.Field(i).Tag.Get("mapstructure"); key != "" {
			keys = append(keys, key)
		}
	}

	return keys
}
