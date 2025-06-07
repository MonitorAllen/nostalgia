package util

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment           string        `mapstructure:"ENVIRONMENT"`
	AllowedOrigins        []string      `mapstructure:"ALLOWED_ORIGINS"`
	DBDriver              string        `mapstructure:"DB_DRIVER"`
	DBUser                string        `mapstructure:"DB_USER"`
	DBPassword            string        `mapstructure:"DB_PASSWORD"`
	DBSource              string        `mapstructure:"DB_SOURCE"`
	MigrationURL          string        `mapstructure:"MIGRATION_URL"`
	ResourcePath          string        `mapstructure:"RESOURCE_PATH"`
	Domain                string        `mapstructure:"DOMAIN"`
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
	HTTPProxyAddr         string        `mapstructure:"HTTP_PROXY_ADDR"`
	DefaultUserID         string        `mapstructure:"DEFAULT_USER_ID"`
	DefaultUsername       string        `mapstructure:"DEFAULT_USERNAME"`
	DefaultUserPassword   string        `mapstructure:"DEFAULT_USER_PASSWORD"`
	DefaultUserFullname   string        `mapstructure:"DEFAULT_USER_FULLNAME"`
	DefaultUserEmail      string        `mapstructure:"DEFAULT_USER_EMAIL"`
}

func LoadConfig(path string) (config Config, err error) {

	viper.SetConfigFile(path + ".env")

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	viper.AutomaticEnv()

	err = viper.Unmarshal(&config)
	return
}
