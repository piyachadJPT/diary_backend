package controllers

import (
	"gofiber-auth/database"
	"gofiber-auth/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetMoodByAdvisor(c *fiber.Ctx) error {
	advisorIDStr := c.Query("advisor_id")
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")

	if advisorIDStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "AdvisorID is required"})
	}

	var startDate, endDate time.Time
	var err error
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid startDate format"})
		}
	}
	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid endDate format"})
		}
	}

	var studentAdvisors []models.StudentAdvisor
	if err := database.DB.Where("advisor_id = ?", advisorIDStr).Find(&studentAdvisors).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if len(studentAdvisors) == 0 {
		statusesList := []string{"veryHappy", "happy", "neutral", "stressed", "burnedOut"}
		counts := make(map[string]int)
		for _, s := range statusesList {
			counts[s] = 0
		}
		return c.JSON(counts)
	}

	studentIDs := make([]uint, len(studentAdvisors))
	for i, sa := range studentAdvisors {
		studentIDs[i] = sa.StudentID
	}

	dbQuery := database.DB.Model(&models.Diary{}).Where("student_id IN ?", studentIDs)
	if !startDate.IsZero() {
		dbQuery = dbQuery.Where("diary_date >= ?", startDate)
	}
	if !endDate.IsZero() {
		dbQuery = dbQuery.Where("diary_date <= ?", endDate)
	}

	var statuses []string
	if err := dbQuery.Pluck("status", &statuses).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	statusesList := []string{"veryHappy", "happy", "neutral", "stressed", "burnedOut"}
	counts := make(map[string]int)
	for _, s := range statusesList {
		counts[s] = 0
	}

	for _, status := range statuses {
		counts[status]++
	}

	return c.JSON(counts)
}
