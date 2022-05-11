package util

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver     string `mapstructure:"DB_DRIVER"`
	DBConnection string `mapstructure:"DB_CONNECTION"`
	DBHost       string `mapstructure:"DB_HOST"`
	DBPort       string `mapstructure:"DB_PORT"`
	DBDatabase   string `mapstructure:"DB_DATABASE"`
	DBUsername   string `mapstructure:"DB_USERNAME"`
	DBPassword   string `mapstructure:"DB_PASSWORD"`

	AppAddress          string        `mapstructure:"APP_ADDRESS"`
	AppPort             string        `mapstructure:"APP_PORT"`
	TokenSymmetricKey   string        `mapstructure:"TOKEN_SYMETRIC_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
}

// Read configuration from file or environtment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

func (c *Config) GetDBSource() string {
	return fmt.Sprintf(
		"%s://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBConnection,
		c.DBUsername,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBDatabase,
	)
}

func (c *Config) GetAppURL() string {
	return fmt.Sprintf("%s:%s", c.AppAddress, c.AppPort)
}
