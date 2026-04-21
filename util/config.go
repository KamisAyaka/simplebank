package util

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	MigrationURL         string        `mapstructure:"MIGRATION_URL"`
	HTTPServerAddress    string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	GRPCServerAddress    string        `mapstructure:"GRPC_SERVER_ADDRESS"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

func LoadConfig(path string) (config Config, err error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("app")
	v.SetConfigType("env")

	v.AutomaticEnv()
	for _, key := range []string{
		"DB_DRIVER",
		"DB_SOURCE",
		"MIGRATION_URL",
		"HTTP_SERVER_ADDRESS",
		"GRPC_SERVER_ADDRESS",
		"TOKEN_SYMMETRIC_KEY",
		"ACCESS_TOKEN_DURATION",
		"REFRESH_TOKEN_DURATION",
	} {
		if bindErr := v.BindEnv(key); bindErr != nil {
			return config, bindErr
		}
	}

	err = v.ReadInConfig()
	if err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		// 允许没有 app.env：此时改从环境变量读取配置。
		if !errors.As(err, &notFoundErr) {
			return
		}
	}
	config = Config{
		DBDriver:             v.GetString("DB_DRIVER"),
		DBSource:             v.GetString("DB_SOURCE"),
		MigrationURL:         v.GetString("MIGRATION_URL"),
		HTTPServerAddress:    v.GetString("HTTP_SERVER_ADDRESS"),
		GRPCServerAddress:    v.GetString("GRPC_SERVER_ADDRESS"),
		TokenSymmetricKey:    v.GetString("TOKEN_SYMMETRIC_KEY"),
		AccessTokenDuration:  v.GetDuration("ACCESS_TOKEN_DURATION"),
		RefreshTokenDuration: v.GetDuration("REFRESH_TOKEN_DURATION"),
	}

	return config, nil
}
