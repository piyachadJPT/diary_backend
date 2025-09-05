package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func MoodRouter(app *fiber.App) {
	app.Get("/api/mood", controllers.GetMoodByAdvisor)
}
