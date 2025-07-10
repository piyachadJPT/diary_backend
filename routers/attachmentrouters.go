package routers

import (
	"gofiber-auth/controllers"

	"github.com/gofiber/fiber/v2"
)

func AttachmentRouter(app *fiber.App) {
	app.Post("/api/diary/uploadfile", controllers.UploadAttachment)
}
