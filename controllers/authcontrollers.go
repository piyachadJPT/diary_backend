package controllers

import (
	"gofiber-auth/database"
	"gofiber-auth/models"

	"github.com/gofiber/fiber/v2"
)

type LoginData struct {
	Email string  `json:"email"`
	Name  *string `json:"name"`
	Image *string `json:"image"`
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
			"error": "Email is required",
		})
	}

	var user models.User
	result := database.DB.Where("email = ?", data.Email).First(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	updated := false

	if data.Name != nil && (user.Name == nil || *user.Name != *data.Name) {
		user.Name = data.Name
		updated = true
	}
	if data.Image != nil && (user.Image == nil || *user.Image != *data.Image) {
		user.Image = data.Image
		updated = true
	}

	if updated {
		if err := database.DB.Save(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update user",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"role":    user.Role,
	})
}
