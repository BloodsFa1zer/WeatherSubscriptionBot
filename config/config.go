package config

import (
	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	URL    string `env:"URL_WEATHER"`
	APIKey string `env:"API_KEY_WEATHER"`
	URI_BD string `env:"URI_MongoDB"`
	BotAPI string `env:"BOT_API_KEY"`
}

func LoadENV(filename string) *Config {
	err := godotenv.Load(filename)
	if err != nil {
		log.Panic().Err(err).Msg(" does not load .env")
	}
	log.Info().Msg("successfully load .env")
	cfg := Config{}
	return &cfg
}

func (cfg *Config) ParseENV() {
	err := env.Parse(cfg)
	if err != nil {
		log.Panic().Err(err).Msg(" unable to parse environment variables")
	}
	log.Info().Msg("successfully parsed .env")

}
