package config

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Mode    string `mapstructure:"mode"`
	Version int    `mapstructure:"version"`
}

func (config AppConfig) Validate() error {
	return validation.ValidateStruct(
		&config,
		validation.Field(&config.Version, validation.Required),
		validation.Field(&config.Name, validation.Required),
		validation.Field(&config.Mode, validation.Required, validation.In("development", "production")),
	)
}

type Monitoring struct {
	OtelEndpoint     string `mapstructure:"otel_endpoint"`
	EnableCollectLog bool   `mapstructure:"enable_collect_log"`
	EnableTelemetry  bool   `mapstructure:"enable_telemetry"`
}

func (c Monitoring) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.OtelEndpoint, validation.Required),
	)
}

type Database struct {
	Adapter      string `json:"adapter" mapstructure:"adapter"`
	Encoding     string `json:"encoding" mapstructure:"encoding"`
	Reconnect    bool   `json:"reconnect" mapstructure:"reconnect"`
	DatabaseName string `json:"database_name" mapstructure:"database_name"`
	SSLMode      bool   `json:"ssl_mode" mapstructure:"sslmode"`
	Alias        string `json:"alias" mapstructure:"alias"`
	Pool         int    `json:"pool" mapstructure:"pool"`
	Schema       string `json:"schema" mapstructure:"schema"`
	Host         string `json:"host" mapstructure:"host"`
	Port         int    `json:"port" mapstructure:"port"`
	Username     string `json:"username" mapstructure:"username"`
	Password     string `json:"password" mapstructure:"password"`
}

func (config Database) Validate() error {
	return validation.ValidateStruct(
		&config,
		validation.Field(&config.Adapter, validation.Required),
		validation.Field(&config.Encoding, validation.Required),
		validation.Field(&config.DatabaseName, validation.Required),
		validation.Field(&config.SSLMode),
		validation.Field(&config.Schema),
		validation.Field(&config.Alias, validation.Required),
		validation.Field(&config.Pool, validation.Required),
		validation.Field(&config.Port, validation.Required),
		validation.Field(&config.Host, validation.Required),
		validation.Field(&config.Username, validation.Required),
		validation.Field(&config.Password, validation.Required),
	)
}

type CorsConfig struct {
	AllowedOrigins []string `json:"allowed_origins" mapstructure:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods" mapstructure:"allowed_methods"`
}

func (config CorsConfig) Validate() error {
	return validation.ValidateStruct(
		&config,
		validation.Field(&config.AllowedOrigins, validation.Required),
		validation.Field(&config.AllowedMethods, validation.Required),
	)
}

type ApiConfig struct {
	Port int        `json:"port" mapstructure:"port"`
	Cors CorsConfig `json:"cors" mapstructure:"cors"`
}

func (config ApiConfig) Validate() error {
	return validation.ValidateStruct(
		&config,
		validation.Field(&config.Port, validation.Required),
		validation.Field(&config.Cors, validation.Required),
	)
}

type Config struct {
	App               AppConfig  `json:"app" mapstructure:"app"`
	Database          Database   `json:"database" mapstructure:"default_database"`
	Monitoring        Monitoring `json:"monitoring" mapstructure:"monitoring"`
	ExternalDatabases []Database `json:"external_databases" mapstructure:"databases_sql"`
	Api               ApiConfig  `json:"apis" mapstructure:"apis"`
}

func NewDefaultConfig() *Config {
	return NewConfig(ConfigParam{
		ConfigName: "config",
		ConfigType: "toml",
		Path:       ".",
	})
}

type ConfigParam struct {
	ConfigName string
	ConfigType string
	Path       string
}

func NewConfig(param ConfigParam) *Config {
	viper.SetConfigName(param.ConfigName)
	viper.SetConfigType(param.ConfigType)

	if param.Path != "" {
		viper.AddConfigPath(param.Path)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic("config not found")
		} else {
			panic(fmt.Errorf("Config error occured, %s", err.Error()))
		}
	}

	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("error while marshalling to struct %s", err.Error()))
	}

	if err := config.Validate(); err != nil {
		panic(err)
	}

	return &config
}

func (config *Config) Validate() error {
	return validation.ValidateStruct(
		config,
		validation.Field(&config.App),
		validation.Field(&config.Database),
		validation.Field(&config.ExternalDatabases),
		validation.Field(&config.Api),
	)
}
