package util

import (
	"errors"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

func LoadConfig(path string) (config Config, err error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("app")
	v.SetConfigType("env")

	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		// 允许没有 app.env：此时改从环境变量读取配置。
		if !errors.As(err, &notFoundErr) {
			return
		}
	}
	err = v.Unmarshal(&config)
	return
}
