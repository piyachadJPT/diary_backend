package main

import (
	"gofiber-auth/database"
	"gofiber-auth/handlers"
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

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	app.Post("/api/auth/microsoft", handlers.HandleMicrosoftLogin)
	app.Get("/api/user/:id", handlers.GetUser)
	app.Post("/api/user", handlers.CreateUser)

	port := os.Getenv("PORT")
	if port == "" {
		port = "6001"
	}

	log.Printf("Starting server at :%s\n", port)
	log.Fatal(app.Listen(":" + port))
}
