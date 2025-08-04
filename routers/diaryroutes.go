package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func DiaryRouter(app *fiber.App) {
	diaryRouter := app.Group("")

	diaryRouter.Post("/api/diary", controllers.CreateNewDiary)
	diaryRouter.Get("/api/diary/", controllers.GetDiaryByDate)
	diaryRouter.Get("/api/diary/by-student", controllers.GetDiaryDateByStudentId)
	diaryRouter.Patch("/api/diary/:id", controllers.PatchDiary)
	diaryRouter.Put("/api/diary/:id", controllers.UpdateDiary)
	diaryRouter.Delete("/api/diary/:id", controllers.DeleteDiary)
}
