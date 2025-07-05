package controllers

import (
	"gofiber-auth/database"
	"gofiber-auth/models"

	"github.com/gofiber/fiber/v2"
)

type LoginData struct {
	Email string `json:"email"`
}

func HandleMicrosoftLogin(c *fiber.Ctx) error {
	var data LoginData
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	if data.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and Name are required",
		})
	}

	var user models.User
	result := database.DB.Where("email = ?", data.Email).First(&user)

	if result.Error != nil {
		user = models.User{
			Email: data.Email,
		}
		createErr := database.DB.Create(&user).Error
		if createErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"role":    user.Role,
	})
}
