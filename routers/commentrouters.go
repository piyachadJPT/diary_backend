package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func CommentRouter(app *fiber.App) {
	app.Post("/api/comment", controllers.CreateNewComment)
	app.Get("/api/comment", controllers.GetCommentByDiaryId)
	app.Delete("/api/comment/:id", controllers.DeleteComment)
}
