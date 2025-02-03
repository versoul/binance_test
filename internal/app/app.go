package app

import (
	"binance_test/internal/config"
	httpController "binance_test/internal/controller/http"
	"binance_test/internal/controller/ws"
	uc "binance_test/internal/usecase"
	"binance_test/internal/ws_server"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

var notify chan error // ошибки сервера

// Run запускает приложение со всеми необходимыми сервисами.
func Run(conf *config.AppConfig) {
	wsServer := ws_server.CreateServer()

	usecases := uc.CreateUsecases(conf, wsServer)
	err := usecases.Binance.Start()
	if err != nil {
		log.Error().Err(err).Msg("Error during startup")
		return
	}

	server := CreateServer(conf, wsServer, usecases)
	log.Info().Msgf("Starting listening on %s", conf.Http.Listen)

	defer func() {
		if err := server.Shutdown(); err != nil {
			notify <- err
		}
	}()

	// Высвободить ресурсы, закрыть соединенения п пр. перед остановкой самого сервера
	terminate := func() {
		usecases.Binance.Stop()
		log.Info().Msg("Cleaning up before terminating the server...")
	}

	sigHandler := make(chan os.Signal, 1)
	signal.Notify(sigHandler, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-sigHandler:
		terminate()
		log.Info().Msg("Everything is cleaned up, stopping the server...")
	case err := <-notify:
		terminate()
		log.Error().Err(err).Msg("Error during startup")

		return
	}
}

// CreateServer создаёт http сервер с необходимыми для приложения middleware и роутами.
func CreateServer(
	conf *config.AppConfig,
	wsServer *ws_server.Server,
	usecases *uc.Usecases,
) *fiber.App {
	fiberConfig := fiber.Config{
		DisableStartupMessage: true,
	}

	app := fiber.New(fiberConfig)

	httpController.Register(app)
	ws.Register(app, wsServer, usecases)

	go func() {
		if err := app.Listen(conf.Http.Listen); err != nil {
			notify <- err
		}
	}()

	return app
}
