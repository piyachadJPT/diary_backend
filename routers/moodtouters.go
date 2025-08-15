package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func MoodRouter(app *fiber.App) {
	moodRouter := app.Group("")

	moodRouter.Get("/api/mood", controllers.GetMoodByAdvisor)
}
