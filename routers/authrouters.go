package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func AuthRouter(app *fiber.App) {
	app.Post("/api/auth/microsoft", controllers.HandleMicrosoftLogin)
	app.Post("/api/auth", controllers.LoginHandler)
}
