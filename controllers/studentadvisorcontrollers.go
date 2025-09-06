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

func CreateStudentAdvisor(c *fiber.Ctx) error {
	var body struct {
		StudentID uint `json:"student_id"`
		AdvisorID uint `json:"advisor_id"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// เช็คว่าคำขอซ้ำอยู่แล้วหรือไม่
	var existing models.AdvisorNotification
	err := database.DB.Where("student_id = ? AND advisor_id = ?", body.StudentID, body.AdvisorID).First(&existing).Error
	if err == nil {
		return c.Status(400).JSON(fiber.Map{
			"ok":      false,
			"message": "คุณได้ส่งคำขออาจารย์ท่านนี้แล้ว",
		})
	}

	var studentAdvisor models.StudentAdvisor
	err = database.DB.Where("student_id = ? AND advisor_id = ?", body.StudentID, body.AdvisorID).First(&studentAdvisor).Error
	if err == nil {
		return c.Status(400).JSON(fiber.Map{
			"ok":      false,
			"message": "คุณเป็นนิสิตในการดูแลอาจารย์ท่านนี้แล้ว",
		})
	}

	// สร้างคำขอในตาราง AdvisorNotification
	notification := models.AdvisorNotification{
		AdvisorID: body.AdvisorID,
		StudentID: body.StudentID,
		Message:   "นิสิตส่งคำร้องขอเป็นที่ปรึกษา",
	}

	if err := database.DB.Create(&notification).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"ok": true,
	})
}

func ApproveAdvisorRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	var notif models.AdvisorNotification
	if err := database.DB.First(&notif, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Request not found"})
	}

	// เพิ่ม StudentAdvisor
	studentAdvisor := models.StudentAdvisor{
		StudentID: notif.StudentID,
		AdvisorID: notif.AdvisorID,
	}
	if err := database.DB.Create(&studentAdvisor).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if err := database.DB.Delete(&notif).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"ok": true, "message": "อนุมัติคำขอเรียบร้อย"})
}

func UnApproveAdvisorRequest(c *fiber.Ctx) error {
	id := c.Params("id")
	var notif models.AdvisorNotification

	if err := database.DB.First(&notif, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Request not found"})
	}

	if err := database.DB.Delete(&notif).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})

	}

	return c.JSON(fiber.Map{"ok": true})
}

func GetAdvisorRequests(c *fiber.Ctx) error {
	advisorID := c.Query("advisor_id")
	if advisorID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "advisor_id required"})
	}

	var requests []models.AdvisorNotification
	if err := database.DB.
		Preload("Student").
		Where("advisor_id = ? AND is_read = ?", advisorID, false).
		Order("created_at desc").
		Find(&requests).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": requests})
}
