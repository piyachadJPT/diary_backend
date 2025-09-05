package main

import (
	"gofiber-auth/database"
	"gofiber-auth/routers"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	database.Connect()

	app := fiber.New(fiber.Config{
		ReadBufferSize:    16384,
		WriteBufferSize:   16384,
		ReadTimeout:       time.Second * 30,
		WriteTimeout:      time.Second * 30,
		IdleTimeout:       time.Minute * 5,
		DisableKeepalive:  false,
		StreamRequestBody: true,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "${time} ${status} - ${method} ${path} - ${latency}\n",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("CORS_ALLOW_ORIGINS"),
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS,HEAD",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With,Cache-Control",
		ExposeHeaders:    "Content-Type,Cache-Control",
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	app.Static("/api/files", "./upload/diary")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Go Fiber Server is running!")
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	routers.UserRouter(app)
	routers.AuthRouter(app)
	routers.DiaryRouter(app)
	routers.AttachmentRouter(app)
	routers.CommentRouter(app)
	routers.StudentRouter(app)
	routers.StudentAdvisorRouter(app)
	routers.MoodRouter(app)
	routers.NotificationRouters(app)
	routers.GroupRouter(app)
	routers.AdminRouter(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("SERVER_PORT")
	}

	log.Printf("Starting server at :%s\n", port)
	log.Fatal(app.Listen(":" + port))
}
