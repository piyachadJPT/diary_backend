package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func DiaryRouter(app *fiber.App) {
	app.Post("/api/diary", controllers.CreateNewDiary)
	app.Get("/api/diary/", controllers.GetDiaryByDate)
	app.Get("/api/diary/by-student", controllers.GetDiaryDateByStudentId)
	app.Patch("/api/diary/:id", controllers.PatchDiary)
	app.Put("/api/diary/:id", controllers.UpdateDiary)
	app.Delete("/api/diary/:id", controllers.DeleteDiary)
}
