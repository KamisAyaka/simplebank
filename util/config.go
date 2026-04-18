package util

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver            string        `mapstructure:"DB_DRIVER"`
	DBSource            string        `mapstructure:"DB_SOURCE"`
	ServerAddress       string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey   string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
}

func LoadConfig(path string) (config Config, err error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("app")
	v.SetConfigType("env")

	v.AutomaticEnv()
	for _, key := range []string{"DB_DRIVER", "DB_SOURCE", "SERVER_ADDRESS"} {
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
		DBDriver:      v.GetString("DB_DRIVER"),
		DBSource:      v.GetString("DB_SOURCE"),
		ServerAddress: v.GetString("SERVER_ADDRESS"),
	}
	if config.DBDriver == "" || config.DBSource == "" || config.ServerAddress == "" {
		return config, fmt.Errorf("missing required config: DB_DRIVER/DB_SOURCE/SERVER_ADDRESS")
	}
	return config, nil
}
