package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

type Config struct {
	AppEnv   string         `mapstructure:"app_env" json:"app_env"`
	GinMode  string         `mapstructure:"gin_mode" json:"gin_mode"`
	Server   ServerConfig   `mapstructure:"server" json:"server"`
	Context  ContextConfig  `mapstructure:"context" json:"context"`
	Database DatabaseConfig `mapstructure:"database" json:"database"`
	JWT      JWTConfig      `mapstructure:"jwt" json:"jwt"`
	Google   GoogleConfig   `mapstructure:"google" json:"google"`
	Log      LogConfig      `mapstructure:"log" json:"log"`
	Email    EmailConfig    `mapstructure:"email" json:"email"`
	OTP      OTPConfig      `mapstructure:"otp" json:"otp"`
}



type ServerConfig struct {
	Address         string        `mapstructure:"address" json:"address"`
	Port            string        `mapstructure:"port" json:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" json:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" json:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" json:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout"`
}

func (config ServerConfig) ListenAddress() string {
	if strings.TrimSpace(config.Address) != "" {
		return config.Address
	}
	port := strings.TrimSpace(config.Port)
	if port == "" {
		port = "8080"
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

type ContextConfig struct {
	Timeout time.Duration `mapstructure:"timeout" json:"timeout"`
}

type DatabaseConfig struct {
	Driver         string        `mapstructure:"driver" json:"driver"`
	URI            string        `mapstructure:"uri" json:"uri"`
	Name           string        `mapstructure:"name" json:"name"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" json:"connect_timeout"`
	PostgresDSN    string        `mapstructure:"postgres_dsn" json:"postgres_dsn"`
}

type JWTConfig struct {
	Secret    string        `mapstructure:"secret" json:"secret"`
	Issuer    string        `mapstructure:"issuer" json:"issuer"`
	ExpiresIn time.Duration `mapstructure:"expires_in" json:"expires_in"`
}

type GoogleConfig struct {
	ClientID     string        `mapstructure:"client_id" json:"client_id"`
	ClientSecret string        `mapstructure:"client_secret" json:"client_secret"`
	RedirectURL  string        `mapstructure:"redirect_url" json:"redirect_url"`
	StateTTL     time.Duration `mapstructure:"state_ttl" json:"state_ttl"`
}

type LogConfig struct {
	Format    string `mapstructure:"format" json:"format"`
	FilePath  string `mapstructure:"file_path" json:"file_path"`
	Level     string `mapstructure:"level" json:"level"`
	ToStdout  bool   `mapstructure:"to_stdout" json:"to_stdout"`
	AddSource bool   `mapstructure:"add_source" json:"add_source"`
}

type EmailConfig struct {
	SMTPHost     string `mapstructure:"smtp_host" json:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port" json:"smtp_port"`
	SMTPUser     string `mapstructure:"smtp_user" json:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password" json:"smtp_password"`
	FromAddress  string `mapstructure:"from_address" json:"from_address"`
	FromName     string `mapstructure:"from_name" json:"from_name"`
}

type OTPConfig struct {
	Length    int           `mapstructure:"length" json:"length"`
	ExpiresIn time.Duration `mapstructure:"expires_in" json:"expires_in"`
}

// envBinding maps a config key to its environment variable name.
type envBinding struct {
	key     string
	envVars []string
}

var envBindings = []envBinding{
	bind("app_env", "APP_ENV"),
	bind("gin_mode", "GIN_MODE"),

	bind("server.address", "SERVER_ADDRESS"),
	bind("server.port", "SERVER_PORT"),
	bind("server.read_timeout", "SERVER_READ_TIMEOUT", "HTTP_READ_TIMEOUT"),
	bind("server.write_timeout", "SERVER_WRITE_TIMEOUT", "HTTP_WRITE_TIMEOUT"),
	bind("server.idle_timeout", "SERVER_IDLE_TIMEOUT", "HTTP_IDLE_TIMEOUT"),
	bind("server.shutdown_timeout", "SERVER_SHUTDOWN_TIMEOUT"),

	bind("context.timeout", "CONTEXT_TIMEOUT"),

	bind("database.driver", "DB_DRIVER"),
	bind("database.uri", "MONGODB_URI"),
	bind("database.name", "DB_NAME"),
	bind("database.connect_timeout", "CONNECT_TIMEOUT"),
	bind("database.postgres_dsn", "POSTGRES_DSN"),

	bind("jwt.secret", "JWT_SECRET"),
	bind("jwt.issuer", "JWT_ISSUER"),
	bind("jwt.expires_in", "JWT_EXPIRES_IN"),

	bind("google.client_id", "GOOGLE_CLIENT_ID"),
	bind("google.client_secret", "GOOGLE_CLIENT_SECRET"),
	bind("google.redirect_url", "GOOGLE_REDIRECT_URL"),
	bind("google.state_ttl", "GOOGLE_STATE_TTL"),

	bind("log.format", "LOG_FORMAT"),
	bind("log.file_path", "LOG_FILE_PATH"),
	bind("log.level", "LOG_LEVEL"),
	bind("log.to_stdout", "LOG_TO_STDOUT"),
	bind("log.add_source", "LOG_ADD_SOURCE"),

	bind("email.smtp_host", "SMTP_HOST"),
	bind("email.smtp_port", "SMTP_PORT"),
	bind("email.smtp_user", "SMTP_USER"),
	bind("email.smtp_password", "SMTP_PASSWORD"),
	bind("email.from_address", "EMAIL_FROM_ADDRESS"),
	bind("email.from_name", "EMAIL_FROM_NAME"),

	bind("otp.length", "OTP_LENGTH"),
	bind("otp.expires_in", "OTP_EXPIRES_IN"),
}

func bind(key string, envVars ...string) envBinding {
	return envBinding{key: key, envVars: envVars}
}

func LoadConfig(path, env string) (Config, error) {
	viper.Reset()

	if err := loadDotEnv(path); err != nil {
		return Config{}, err
	}

	setDefaults()
	bindEnvKeys()

	if env == "" {
		env = os.Getenv("ENV")
	}
	if env == "" {
		env = "development"
	}

	configFile := filepath.Join(path, fmt.Sprintf("%s.yaml", env))
	if _, statErr := os.Stat(configFile); statErr == nil {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			return Config{}, err
		}
	} else if !os.IsNotExist(statErr) {
		return Config{}, statErr
	}

	var config Config
	if err := viper.Unmarshal(&config, viper.DecodeHook(timeDurationHookFunc())); err != nil {
		return Config{}, err
	}
	if config.Log.AddSource || env == "development" {
		config.Log.AddSource = true
	}

	if err := validate(config); err != nil {
		return Config{}, err
	}

	if env == "development" {
		b, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return Config{}, err
		}
		fmt.Println(string(b))
	}

	return config, nil
}

func loadDotEnv(path string) error {
	if path == "" {
		path = "."
	}

	envFile := filepath.Join(path, ".env")
	if _, err := os.Stat(envFile); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := gotenv.Load(envFile); err != nil {
		return fmt.Errorf("load .env: %w", err)
	}
	return nil
}

func validate(config Config) error {
	if len(config.JWT.Secret) < 16 {
		return fmt.Errorf("jwt.secret must be at least 16 characters")
	}
	return nil
}

func setDefaults() {
	viper.SetDefault("app_env", "development")
	viper.SetDefault("gin_mode", "debug")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", "5s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.idle_timeout", "1m")
	viper.SetDefault("server.shutdown_timeout", "10s")
	viper.SetDefault("context.timeout", "2s")
	viper.SetDefault("database.driver", "mongo")
	viper.SetDefault("database.uri", "mongodb://localhost:27017")
	viper.SetDefault("database.name", "todo_app")
	viper.SetDefault("database.connect_timeout", "10s")
	viper.SetDefault("database.postgres_dsn", "postgres://postgres@localhost:5432/postgres?sslmode=disable")
	viper.SetDefault("jwt.issuer", "go-crud-db-p2")
	viper.SetDefault("jwt.expires_in", "24h")
	viper.SetDefault("google.redirect_url", "http://localhost:8080/api/v1/auth/google/callback")
	viper.SetDefault("google.state_ttl", "10m")
	viper.SetDefault("log.format", "text")
	viper.SetDefault("log.file_path", "storage/logs/app.log")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.to_stdout", true)

	viper.SetDefault("email.smtp_host", "")
	viper.SetDefault("email.smtp_port", 587)
	viper.SetDefault("email.from_name", "Todo App")

	viper.SetDefault("otp.length", 6)
	viper.SetDefault("otp.expires_in", "10m")
}

// bindEnvKeys explicitly binds all config keys to environment variables.
// This is required for env vars to override YAML values during Unmarshal().
func bindEnvKeys() {
	for _, binding := range envBindings {
		envVars := append([]string{binding.key}, binding.envVars...)
		_ = viper.BindEnv(envVars...)
	}
}

func timeDurationHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Duration(0)) {
			return data, nil
		}
		switch f.Kind() {
		case reflect.String:
			s := data.(string)
			if d, err := time.ParseDuration(s); err == nil {
				return d, nil
			}
			ns, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse %q as duration: expected format like \"24h\" or nanoseconds", s)
			}
			return time.Duration(ns), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return time.Duration(reflect.ValueOf(data).Int()), nil
		case reflect.Float32, reflect.Float64:
			return time.Duration(int64(reflect.ValueOf(data).Float())), nil
		default:
			return data, nil
		}
	}
}
