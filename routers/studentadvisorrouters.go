package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func StudentAdvisorRouter(app *fiber.App) {
	studentAdvisorRouter := app.Group("")

	studentAdvisorRouter.Get("/api/studentAdvisor", controllers.GetStudentByAdvisor)
	studentAdvisorRouter.Delete("/api/studentAdvisor", controllers.DeleteStudentAdvisor)
}
