package handlers

import (
	"gofiber-auth/database"
	"gofiber-auth/models"

	"github.com/gofiber/fiber/v2"
)

func GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User
	result := database.DB.First(&user, id)
	if result.Error != nil {
		c.Status(404)
		return c.SendString("User not found")
	}
	return c.JSON(user)
}

func CreateUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		c.Status(400)
		return c.SendString("Cannot parse JSON")
	}

	database.DB.Create(&user)

	return c.SendString("User created successfully")
}
