package controllers

import (
	"gofiber-auth/database"
	"gofiber-auth/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetStudentByAdvisor(c *fiber.Ctx) error {
	advisorIDStr := c.Query("advisor_id")

	if advisorIDStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Student advisor are required",
		})
	}

	var advisor []models.StudentAdvisor

	result := database.DB.
		Preload("Student").
		Where("advisor_id = ?", advisorIDStr).
		Find(&advisor)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	if len(advisor) == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "No student entries found for this advisor_id",
		})
	}

	var resultList []fiber.Map
	for _, a := range advisor {
		resultList = append(resultList, fiber.Map{
			"id":         a.ID,
			"student_id": a.StudentID,
			"AdvisorID":  a.AdvisorID,
			"student":    a.Student,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Advisor retrieved successfully",
		"data":    resultList,
	})
}

func DeleteStudentAdvisor(c *fiber.Ctx) error {
	idParam := c.Query("id")
	if idParam == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Student advisor ID is required",
		})
	}

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	result := database.DB.Delete(&models.StudentAdvisor{}, id)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "Student advisor not found",
		})
	}

	return c.JSON(fiber.Map{
		"ok":      true,
		"message": "Advisor deleted successfully",
	})
}
