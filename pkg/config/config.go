package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	NatsURL     string `mapstructure:"NATS_URL"`
	PluginDir   string `mapstructure:"PLUGIN_DIR"`
	Port        string `mapstructure:"PORT"`
}

func Load() (*Config, error) {
	viper.SetDefault("NATS_URL", "nats://localhost:4222")
	viper.SetDefault("PLUGIN_DIR", "./plugins")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DATABASE_URL", "postgres://webencode:webencode@localhost:5432/webencode?sslmode=disable")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Check for .env file
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Warning: Config file not found: %v", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
