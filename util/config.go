package util

import (
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Environment           string        `mapstructure:"ENVIRONMENT"`
	AllowedOrigins        []string      `mapstructure:"ALLOWED_ORIGINS"`
	DBDriver              string        `mapstructure:"DB_DRIVER"`
	DBSource              string        `mapstructure:"DB_SOURCE"`
	MigrationURL          string        `mapstructure:"MIGRATION_URL"`
	RedisAddress          string        `mapstructure:"REDIS_ADDRESS"`
	HTTPServerAddress     string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	GrpcGatewayAddress    string        `mapstructure:"GRPC_GATEWAY_ADDRESS"`
	GRPCServerAddress     string        `mapstructure:"GRPC_SERVER_ADDRESS"`
	TokenSymmetricKey     string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration   time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration  time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	EmailSenderName       string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress    string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword   string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
	UploadFileSizeLimit   int64         `mapstructure:"UPLOAD_FILE_SIZE_LIMIT"`
	UploadFileAllowedMime []string      `mapstructure:"UPLOAD_FILE_ALLOWED_MIME"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	viper.AutomaticEnv()

	err = viper.Unmarshal(&config)
	return
}
