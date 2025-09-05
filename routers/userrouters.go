package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func UserRouter(app *fiber.App) {
	app.Get("/api/user/:id", controllers.GetUser)
	app.Get("/api/user", controllers.GetUserByEmail)
	app.Get("/api/alluser/", controllers.GetAllUserNotconfirmed)
	app.Post("/api/user", controllers.CreateUser)
	app.Get("/api/profile", controllers.AuthMiddleware, controllers.GetProfileHandler)
	app.Patch("/api/user/:id", controllers.PatchApproved)
	app.Delete("/api/user/:id", controllers.DeleteApprover)
	app.Post("/api/user/register", controllers.RegisterHandler)
}
