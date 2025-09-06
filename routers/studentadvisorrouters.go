package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func StudentAdvisorRouter(app *fiber.App) {
	app.Get("/api/studentAdvisor", controllers.GetStudentByAdvisor)
	app.Delete("/api/studentAdvisor", controllers.DeleteStudentAdvisor)
	app.Post("/api/student-advisor", controllers.CreateStudentAdvisor)
	app.Patch("/api/student-advisor", controllers.ApproveAdvisorRequest)
	app.Get("/api/student-advisor", controllers.GetAdvisorRequests)
	app.Delete("/api/student-advisor", controllers.UnApproveAdvisorRequest)
}
