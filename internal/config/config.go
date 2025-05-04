package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
	Storage  StorageConfig  `mapstructure:"storage"`
}

// AppConfig holds application specific configuration.
type AppConfig struct {
	DebugMode bool `mapstructure:"debug_mode"`
}

// ServerConfig holds server related configuration.
type ServerConfig struct {
	Address string `mapstructure:"address"`
	BaseURL string `mapstructure:"base_url"` // Base URL for constructing public links
}

// DatabaseConfig holds database connection details.
type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	Source string `mapstructure:"source"`
}

// JWTConfig holds JWT related settings.
type JWTConfig struct {
	Secret              string        `mapstructure:"secret"`
	AccessTokenDuration time.Duration `mapstructure:"access_token_duration"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Enable     bool   `mapstructure:"enable"`
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
	Beautify   bool   `mapstructure:"beautify"`
	Trace      bool   `mapstructure:"trace"`
}

// StorageConfig holds storage related configurations.
type StorageConfig struct {
	Local LocalStorageConfig `mapstructure:"local"`
	// Add S3Config, MinioConfig etc. here later
}

// LocalStorageConfig holds configurations for local filesystem storage.
type LocalStorageConfig struct {
	BasePath string `mapstructure:"base_path"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	viper.AutomaticEnv() // Read environment variables

	// Set default values (optional but recommended)
	viper.SetDefault("app.debug_mode", false)
	viper.SetDefault("server.address", ":8080")
	viper.SetDefault("server.base_url", "http://localhost:8080") // Default base URL
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.source", "memento.db")
	viper.SetDefault("jwt.secret", "change_this_secret")
	viper.SetDefault("jwt.access_token_duration", "15m")
	viper.SetDefault("log.enable", true)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", "")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_age", 7)
	viper.SetDefault("log.max_backups", 3)
	viper.SetDefault("log.compress", false)
	viper.SetDefault("log.beautify", false)
	viper.SetDefault("log.trace", true)
	viper.SetDefault("storage.local.base_path", "assets/images")

	err = viper.ReadInConfig()
	// Ignore 'config file not found' error if defaults are sufficient
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok && err != nil {
		return config, err // Return other errors
	}

	err = viper.Unmarshal(&config)
	return config, err
}
