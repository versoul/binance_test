package ws

import (
	"binance_test/internal/usecase"
	"binance_test/internal/ws_server"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func Register(app *fiber.App, wsServer *ws_server.Server, usecases *usecase.Usecases) {
	// ws соединение
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("id", uuid.New().String())
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/bids", websocket.New(func(c *websocket.Conn) {

		log.Debug().Msg("New WebSocket connection")
		wsServer.AddClient(c.Locals("id").(string), c)
		usecases.Binance.SendDataInitial()

		c.SetCloseHandler(func(code int, text string) error {

			wsServer.RemoveClient(c.Locals("id").(string))
			log.Debug().Msg("Closed WebSocket connection")

			return nil
		})

		wsServer.Read(c)
	}))
}
