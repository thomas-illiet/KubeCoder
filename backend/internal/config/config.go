package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	HTTP     HTTPConfig     `mapstructure:"http"`
	Database DatabaseConfig `mapstructure:"database"`
	OIDC     OIDCConfig     `mapstructure:"oidc"`
	Security SecurityConfig `mapstructure:"security"`
	Log      LogConfig      `mapstructure:"log"`
}

type SecurityConfig struct {
	OrganizationSSHKeyEncryptionKey string `mapstructure:"organization_ssh_key_encryption_key"`
}

type HTTPConfig struct {
	Address         string        `mapstructure:"address"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	AllowedOrigins  []string      `mapstructure:"allowed_origins"`
}

type DatabaseConfig struct {
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type OIDCConfig struct {
	Issuer       string        `mapstructure:"issuer"`
	DiscoveryURL string        `mapstructure:"discovery_url"`
	Audience     string        `mapstructure:"audience"`
	Timeout      time.Duration `mapstructure:"timeout"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Defaults returns safe operational defaults for optional configuration values.
func Defaults() Config {
	return Config{
		HTTP:     HTTPConfig{Address: ":8080", ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, ShutdownTimeout: 15 * time.Second},
		Database: DatabaseConfig{MaxOpenConns: 20, MaxIdleConns: 5, ConnMaxLifetime: 30 * time.Minute},
		OIDC:     OIDCConfig{Timeout: 10 * time.Second},
		Log:      LogConfig{Level: "info", Format: "json"},
	}
}

// Load merges defaults, an optional YAML file, environment variables, and flags.
func Load(configFile string, flags *pflag.FlagSet) (Config, error) {
	defaults := Defaults()
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("KUBECODER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	for _, key := range []string{
		"http.address", "http.read_timeout", "http.write_timeout", "http.idle_timeout", "http.shutdown_timeout", "http.allowed_origins",
		"database.dsn", "database.max_open_conns", "database.max_idle_conns", "database.conn_max_lifetime",
		"oidc.issuer", "oidc.discovery_url", "oidc.audience", "oidc.timeout", "log.level", "log.format",
		"security.organization_ssh_key_encryption_key",
	} {
		if err := v.BindEnv(key); err != nil {
			return Config{}, fmt.Errorf("bind environment variable for %s: %w", key, err)
		}
	}

	setDefaults(v, defaults)
	if flags != nil {
		for _, key := range []string{"http.address", "http.allowed_origins", "database.dsn", "oidc.issuer", "oidc.discovery_url", "oidc.audience", "log.level", "log.format"} {
			if flag := flags.Lookup(key); flag != nil {
				if err := v.BindPFlag(key, flag); err != nil {
					return Config{}, fmt.Errorf("bind flag %s: %w", key, err)
				}
			}
		}
	}
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return Config{}, fmt.Errorf("read configuration: %w", err)
		}
	}

	var cfg Config
	if err := v.UnmarshalExact(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode configuration: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// OrganizationSSHKeyEncryptionKey decodes and validates the required AES-256 master key.
func (c Config) OrganizationSSHKeyEncryptionKey() ([]byte, error) {
	value := strings.TrimSpace(c.Security.OrganizationSSHKeyEncryptionKey)
	key, err := base64.StdEncoding.Strict().DecodeString(value)
	if err != nil || len(key) != 32 {
		return nil, errors.New("security.organization_ssh_key_encryption_key must be base64 encoding exactly 32 bytes")
	}
	return key, nil
}

// setDefaults registers default values with a Viper instance.
func setDefaults(v *viper.Viper, cfg Config) {
	v.SetDefault("http.address", cfg.HTTP.Address)
	v.SetDefault("http.read_timeout", cfg.HTTP.ReadTimeout)
	v.SetDefault("http.write_timeout", cfg.HTTP.WriteTimeout)
	v.SetDefault("http.idle_timeout", cfg.HTTP.IdleTimeout)
	v.SetDefault("http.shutdown_timeout", cfg.HTTP.ShutdownTimeout)
	v.SetDefault("database.max_open_conns", cfg.Database.MaxOpenConns)
	v.SetDefault("database.max_idle_conns", cfg.Database.MaxIdleConns)
	v.SetDefault("database.conn_max_lifetime", cfg.Database.ConnMaxLifetime)
	v.SetDefault("oidc.timeout", cfg.OIDC.Timeout)
	v.SetDefault("log.level", cfg.Log.Level)
	v.SetDefault("log.format", cfg.Log.Format)
}

// Validate checks that the configuration is complete and internally consistent.
func (c Config) Validate() error {
	var errs []error
	if c.Database.DSN == "" {
		errs = append(errs, errors.New("database.dsn is required"))
	}
	if c.OIDC.Issuer == "" {
		errs = append(errs, errors.New("oidc.issuer is required"))
	}
	if c.OIDC.Audience == "" {
		errs = append(errs, errors.New("oidc.audience is required"))
	}
	if c.Database.MaxOpenConns < 1 || c.Database.MaxIdleConns < 0 || c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		errs = append(errs, errors.New("database connection pool limits are invalid"))
	}
	if _, err := ParseLogLevel(c.Log.Level); err != nil {
		errs = append(errs, err)
	}
	if c.Log.Format != "json" && c.Log.Format != "text" {
		errs = append(errs, errors.New("log.format must be json or text"))
	}
	return errors.Join(errs...)
}

// ParseLogLevel converts a textual slog level into its typed representation.
func ParseLogLevel(value string) (slog.Level, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(value)); err != nil {
		return level, fmt.Errorf("invalid log.level %q: %w", value, err)
	}
	return level, nil
}
