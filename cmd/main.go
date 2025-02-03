package main

import (
	"binance_test/internal/app"
	"binance_test/internal/config"
	"github.com/rs/zerolog/log"
)

func main() {
	// Инициализация структуры конфига и парсинг флагов.
	if err := config.Init(); err != nil {
		log.Fatal().Err(err).Msg("Initialization app config failed.")
	}
	defer log.Info().Msg("Exiting normally")

	app.Run(config.Config)
}
