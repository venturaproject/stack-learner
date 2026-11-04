package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Cookie   CookieConfig
	CORS     CORSConfig
	Log      LogConfig
	Media    MediaConfig
	Auth     AuthConfig
}

type AuthConfig struct {
	// FullAccessRoles bypass every permission check (mirrors the frontend
	// VITE_FULL_ACCESS_ROLES).
	FullAccessRoles []string
}

type ServerConfig struct {
	Port int
	Mode string
}

type DatabaseConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret         string
	AccessTTLSecs  int
	RefreshTTLSecs int
}

type CookieConfig struct {
	Secure   bool
	SameSite string
	Domain   string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type LogConfig struct {
	Level string
}

type MediaConfig struct {
	Dir        string
	PrivateDir string
}

func Load() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetDefault("SERVER_PORT", 8000)
	viper.SetDefault("SERVER_MODE", "debug")
	viper.SetDefault("JWT_ACCESS_TTL_SECS", 900)
	viper.SetDefault("JWT_REFRESH_TTL_SECS", 604800)
	viper.SetDefault("COOKIE_SECURE", false)
	viper.SetDefault("COOKIE_SAME_SITE", "Lax")
	viper.SetDefault("REDIS_ADDR", "redis:6379")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("MEDIA_DIR", "./media")
	viper.SetDefault("MEDIA_PRIVATE_DIR", "./media-private")
	viper.SetDefault("FULL_ACCESS_ROLES", "admin")

	cfg := &Config{
		Server: ServerConfig{
			Port: viper.GetInt("SERVER_PORT"),
			Mode: viper.GetString("SERVER_MODE"),
		},
		Database: DatabaseConfig{
			DSN: buildDSN(),
		},
		Redis: RedisConfig{
			Addr:     viper.GetString("REDIS_ADDR"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			Secret:         viper.GetString("JWT_SECRET"),
			AccessTTLSecs:  viper.GetInt("JWT_ACCESS_TTL_SECS"),
			RefreshTTLSecs: viper.GetInt("JWT_REFRESH_TTL_SECS"),
		},
		Cookie: CookieConfig{
			Secure:   viper.GetBool("COOKIE_SECURE"),
			SameSite: viper.GetString("COOKIE_SAME_SITE"),
			Domain:   viper.GetString("COOKIE_DOMAIN"),
		},
		CORS: CORSConfig{
			AllowedOrigins: parseCORSOrigins(viper.GetString("CORS_ALLOWED_ORIGINS")),
		},
		Log: LogConfig{
			Level: viper.GetString("LOG_LEVEL"),
		},
		Media: MediaConfig{
			Dir:        viper.GetString("MEDIA_DIR"),
			PrivateDir: viper.GetString("MEDIA_PRIVATE_DIR"),
		},
		Auth: AuthConfig{
			FullAccessRoles: parseCSV(viper.GetString("FULL_ACCESS_ROLES")),
		},
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func buildDSN() string {
	if dsn := viper.GetString("DATABASE_URL"); dsn != "" {
		return dsn
	}
	host := viper.GetString("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := viper.GetString("DB_PORT")
	if port == "" {
		port = "5432"
	}
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		host, port,
		viper.GetString("DB_USER"),
		viper.GetString("DB_PASSWORD"),
		viper.GetString("DB_NAME"),
	)
}

func parseCORSOrigins(s string) []string {
	if s == "" {
		return []string{"http://localhost:5173"}
	}
	return parseCSV(s)
}

// parseCSV splits a comma-separated env value into trimmed, non-empty parts.
func parseCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
