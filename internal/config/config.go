package config

import "time"

type Config struct {
	Server   Server   `env-prefix:"SERVER_"`
	Postgres Postgres `env-prefix:"POSTGRES_"`
	Telegram Telegram `env-prefix:"TELEGRAM_"`
	Auth     Auth     `env-prefix:"AUTH_"`
	Minio    Minio    `env-prefix:"MINIO_"`
}

type Server struct {
	Host           string   `env:"HOST"`
	Port           int      `env:"PORT"`
	BasePath       string   `env:"BASE_PATH"`
	AllowedOrigins []string `env:"ALLOWED_ORIGINS"`
	AllowedMethods []string `env:"ALLOWED_METHODS"`
	AllowedHeaders []string `env:"ALLOWED_HEADERS"`
}

type Auth struct {
	AccessTokenTTL  time.Duration `env:"ACCESS_TOKEN_TTL"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL"`
	AuthSecret      string        `env:"AUTH_SECRET"`
}

type Postgres struct {
	DSN           string `env:"DSN"`
	Migrate       bool   `env:"MIGRATE"`
	MigrationsDir string `env:"MIGRATIONS_DIR"`
}

type Telegram struct {
	Token string `env:"TOKEN"`
}

type Minio struct {
	Port         int    `env:"PORT"`
	Endpoint     string `env:"ENDPOINT"`
	RootUser     string `env:"ROOT_USER"`
	RootPassword string `env:"ROOT_PASSWORD"`
	UseSSL       bool   `env:"USE_SSL"`
	BucketName   string `env:"BUCKET_NAME"`
}
