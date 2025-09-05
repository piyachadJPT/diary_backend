package routers

import (
	"gofiber-auth/controllers"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NotificationRouters(app *fiber.App) {
	notificationRouters := app.Group("/api/notification")

	notificationRouters.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("CORS_ALLOW_ORIGINS"),
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With,Cache-Control",
		ExposeHeaders:    "Content-Type,Cache-Control",
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	notificationRouters.Options("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(204)
	})

	notificationRouters.Get("/stream", controllers.AdvisorSSE)

	notificationRouters.Get("/unread", controllers.GetUnreadNotifications)
	notificationRouters.Get("/all", controllers.GetAllNotifications)
	notificationRouters.Delete("/:id", controllers.DeleteNotification)

	notificationRouters.Patch("/:id/read", controllers.MarkNotificationAsRead)
	notificationRouters.Patch("/read-all", controllers.MarkAllNotificationsAsRead)
	notificationRouters.Get("/count", controllers.GetNotificationCount)

	notificationRouters.Get("/connections", controllers.GetActiveConnections)
}
