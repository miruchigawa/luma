package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the root configuration.
type Config struct {
	Bot     BotConfig     `mapstructure:"bot" yaml:"bot"`
	Session SessionConfig `mapstructure:"session" yaml:"session"`
	Log     LogConfig     `mapstructure:"log" yaml:"log"`
}

// BotConfig holds bot-specific settings.
type BotConfig struct {
	CommandPrefixes []string `mapstructure:"command_prefixes" yaml:"command_prefixes"`
	ListenSelf      bool     `mapstructure:"listen_self" yaml:"listen_self"`
	AdminJIDs       []string `mapstructure:"admin_jids" yaml:"admin_jids"`
}

// SessionConfig holds session store settings.
type SessionConfig struct {
	Driver      string         `mapstructure:"driver" yaml:"driver"`
	LoginMethod string         `mapstructure:"login_method" yaml:"login_method"`
	PhoneNumber string         `mapstructure:"phone_number" yaml:"phone_number"`
	SQLite      SQLiteConfig   `mapstructure:"sqlite" yaml:"sqlite"`
	Postgres    PostgresConfig `mapstructure:"postgres" yaml:"postgres"`
}

// SQLiteConfig holds SQLite connection details.
type SQLiteConfig struct {
	Path string `mapstructure:"path" yaml:"path"`
}

// PostgresConfig holds PostgreSQL connection details.
type PostgresConfig struct {
	DSN string `mapstructure:"dsn" yaml:"dsn"`
}

// LogConfig holds logger settings.
type LogConfig struct {
	Level string `mapstructure:"level" yaml:"level"`
}

// Load reads the configuration from config.yaml and overrides with environment variables.
//
// Key design decisions:
// - Uses Viper for unified file and ENV loading.
// - ENV variables are prefixed with BOT_ (e.g., BOT_SESSION_DRIVER).
// - Struct tags sync seamlessly between YAML keys and Viper dictionary structure.
func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("config/")

	v.SetEnvPrefix("bot")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Setting default values
	v.SetDefault("bot.command_prefixes", []string{"!"})
	v.SetDefault("bot.listen_self", false)
	v.SetDefault("session.driver", "sqlite")
	v.SetDefault("session.login_method", "qr")
	v.SetDefault("log.level", "info")
	v.SetDefault("session.sqlite.path", "data/session.db")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found; ignore error to allow pure ENV configuration.
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
