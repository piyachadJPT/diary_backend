package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func StudentRouter(app *fiber.App) {
	app.Get("/api/students", controllers.GetAllStudents)
	app.Get("/api/student", controllers.GetStudentById)
	app.Post("/api/student", controllers.CreateStudent)
}
