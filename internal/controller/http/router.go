package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"net/http"
)

func Register(app *fiber.App) {
	// Отображение статики для тестирования
	app.Use("/static", filesystem.New(filesystem.Config{
		Root:   http.Dir("./static"),
		Browse: true,
	}))
}
