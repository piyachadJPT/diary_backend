package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func GroupRouter(app *fiber.App) {
	app.Get("/api/groups", controllers.GetGroupsByAdvisor)
	app.Post("/api/groups", controllers.CreateGroup)
	app.Delete("/api/groups", controllers.DeleteGroup)

	// สำหรับนิสิตที่มีกลุ่ม
	app.Get("/api/groups/students", controllers.GetStudentsInGroup)
	app.Post("/api/groups/students", controllers.AddStudentToGroup)
	app.Delete("/api/groups/students", controllers.RemoveStudentFromGroup)

	// ไม่มีกลุ่ม
	app.Get("/api/students/without-group", controllers.GetStudentsWithoutGroup)
}
