package controllers

import (
	"gofiber-auth/database"
	"gofiber-auth/models"

	"github.com/gofiber/fiber/v2"
)

func GetAllStudentsByAdmin(c *fiber.Ctx) error {
	var students []models.User
	result := database.DB.Where("role = ?", "student").Find(&students)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to query students")
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).SendString("No students found")
	}

	return c.JSON(students)
}

func GetAllTeacherByAdmin(c *fiber.Ctx) error {
	var teachers []models.User
	result := database.DB.Where("role = ?", "advisor").Find(&teachers)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to query teachers")
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).SendString("No teachers found")
	}

	return c.JSON(teachers)
}
