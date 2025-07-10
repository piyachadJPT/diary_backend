package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func CommentRouter(app *fiber.App) {
	commentRouter := app.Group("")

	commentRouter.Post("/api/comment", controllers.CreateNewComment)
	commentRouter.Get("/api/comment", controllers.GetCommentByDiaryId)
	commentRouter.Delete("/api/comment/:id", controllers.DeleteComment)

}
