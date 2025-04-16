package config

type Config struct {
	Server   Server   `env-prefix:"SERVER_"`
	Postgres Postgres `env-prefix:"POSTGRES_"`
}

type Server struct {
	Host           string   `env:"HOST"`
	Port           int      `env:"PORT"`
	BasePath       string   `env:"BASE_PATH"`
	AllowedOrigins []string `env:"ALLOWED_ORIGINS"`
	AllowedMethods []string `env:"ALLOWED_METHODS"`
	AllowedHeaders []string `env:"ALLOWED_HEADERS"`
}

type Postgres struct {
	DSN           string `env:"DSN"`
	Migrate       bool   `env:"MIGRATE"`
	MigrationsDir string `env:"MIGRATIONS_DIR"`
}
