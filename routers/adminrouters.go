package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func AdminRouter(app *fiber.App) {
	app.Get("/api/admin/allstudent", controllers.GetAllStudentsByAdmin)
	app.Get("/api/admin/allteachers", controllers.GetAllTeacherByAdmin)
}
