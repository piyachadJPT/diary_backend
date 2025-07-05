package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func UserRouter(app *fiber.App) {
	app.Get("/api/user/:id", controllers.GetUser)
	app.Get("/api/user", controllers.GetUserByEmail)
	app.Post("/api/user", controllers.CreateUser)
}
