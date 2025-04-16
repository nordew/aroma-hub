package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"sync"
)

var (
	once     sync.Once
	instance Config
)

func MustLoad() Config {
	once.Do(func() {
		if err := cleanenv.ReadEnv(&instance); err != nil {
			log.Fatalf("Error loading environment variables: %v", err)
		}
	})

	return instance
}
