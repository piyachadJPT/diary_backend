package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func StudentRouter(app *fiber.App) {
	studentRouter := app.Group("")

	studentRouter.Get("/api/students", controllers.GetAllStudents)
	studentRouter.Get("/api/student", controllers.GetStudentById)
	// studentRouter.Get("/api/student", controllers.GetStudentByAdvisor)
	studentRouter.Post("/api/student", controllers.CreateStudent)
}
