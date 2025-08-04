package main

import (
	"gofiber-auth/database"
	"gofiber-auth/routers"

	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	database.Connect()
	app := fiber.New(fiber.Config{
		ReadBufferSize: 16384,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("CORS_ALLOW_ORIGINS"),
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	app.Static("/api/files", "./upload/diary")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Go Fiber Server is running!")
	})

	routers.UserRouter(app)
	routers.AuthRouter(app)
	routers.DiaryRouter(app)
	routers.AttachmentRouter(app)
	routers.CommentRouter(app)
	routers.StudentRouter(app)
	routers.StudentAdvisorRouter(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("SERVER_PORT")
	}

	log.Printf("Starting server at :%s\n", port)
	log.Fatal(app.Listen(":" + port))
}
