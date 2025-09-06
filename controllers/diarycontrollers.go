package controllers

import (
	"encoding/json"
	"fmt"
	"gofiber-auth/database"
	"gofiber-auth/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetDiaryByDate(c *fiber.Ctx) error {
	DiaryDate := c.Query("DiaryDate")
	studentID := c.Query("StudentID")

	if DiaryDate == "" || studentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "DiaryDate and StudentID are required",
		})
	}

	var diaries []models.Diary

	result := database.DB.
		Preload("Student").
		Preload("Attachments").
		Where("diary_date = ? AND student_id = ?", DiaryDate, studentID).
		Find(&diaries)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	if len(diaries) == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "No diary entries found for this date and student",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Diaries retrieved successfully",
		"data":    diaries,
	})
}

func CreateNewDiary(c *fiber.Ctx) error {
	var diary models.Diary

	if err := c.BodyParser(&diary); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Cannot parse JSON",
			"details": err.Error(),
		})
	}

	if diary.StudentID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "StudentID is required",
		})
	}

	if diary.ContentHTML == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ContentHTML is required",
		})
	}

	if diary.IsShared == "" {
		diary.IsShared = "everyone"
	}

	if diary.Status == "" {
		diary.Status = "neutral"
	}

	if diary.DiaryDate.IsZero() {
		diary.DiaryDate = time.Now()
	}

	// บันทึก Diary ในฐานข้อมูล
	if err := database.DB.Create(&diary).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create diary",
			"details": err.Error(),
		})
	}

	//  สร้าง Notification สำหรับอาจารย์ที่ปรึกษา
	var advisors []models.StudentAdvisor
	result := database.DB.Preload("Advisor").Where("student_id = ?", diary.StudentID).Find(&advisors)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to find advisors",
			"details": result.Error.Error(),
		})
	}

	for _, sa := range advisors {
		var student models.User
		database.DB.First(&student, diary.StudentID)

		notificationData := map[string]interface{}{
			"diary_date": diary.DiaryDate.Format("2006-01-02"),
			"student_id": diary.StudentID,
		}

		dataJSON, _ := json.Marshal(notificationData)

		var studentName string
		if student.Name != nil {
			studentName = *student.Name
		} else {
			studentName = "ไม่ระบุชื่อ"
		}

		notif := models.Notification{
			UserID:  sa.AdvisorID,
			DiaryID: &diary.ID,
			Type:    "new_diary",
			Title:   "นิสิตเพิ่มบันทึกใหม่",
			Message: fmt.Sprintf("%s เพิ่มข้อมูลใหม่", studentName),
			Data:    dataJSON,
			IsRead:  false,
		}

		if err := database.DB.Create(&notif).Error; err != nil {
			continue
		}

		if conn, ok := advisorChannels[sa.AdvisorID]; ok {
			select {
			case conn.Ch <- notif:
			default:
			}
		}

	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Created diary successfully",
		"data":    diary,
	})
}

func UpdateDiary(c *fiber.Ctx) error {
	id := c.Params("id")
	var diary models.Diary

	if result := database.DB.First(&diary, id); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Diary not found",
		})
	}

	var updateData models.Diary
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Cannot parse JSON",
			"details": err.Error(),
		})
	}

	diary.ContentHTML = updateData.ContentHTML
	diary.ContentDelta = updateData.ContentDelta
	diary.IsShared = updateData.IsShared
	diary.AllowComment = updateData.AllowComment
	diary.Status = updateData.Status

	if err := database.DB.Save(&diary).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to update diary",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Updated diary successfully",
		"data":    diary,
	})
}

func PatchDiary(c *fiber.Ctx) error {
	id := c.Params("id")
	var diary models.Diary

	if result := database.DB.First(&diary, id); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Diary not found",
		})
	}

	var updateData map[string]interface{}
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Cannot parse JSON",
			"details": err.Error(),
		})
	}

	if err := database.DB.Model(&diary).Updates(updateData).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to update diary",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Patched diary successfully",
		"data":    diary,
	})
}

func DeleteDiary(c *fiber.Ctx) error {
	id := c.Params("id")
	var diary models.Diary

	if result := database.DB.First(&diary, id); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Diary not found",
		})
	}

	if err := database.DB.Delete(&diary).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to delete diary",
			"details": err.Error(),
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"ok":      true,
		"message": "Deleted diary successfully",
	})
}

func GetDiariesByStudent(c *fiber.Ctx) error {
	studentID := c.Params("studentId")
	var diaries []models.Diary

	if err := database.DB.Where("student_id = ?", studentID).
		Order("diary_date desc").
		Find(&diaries).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve diaries",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Diaries retrieved successfully",
		"data":    diaries,
		"count":   len(diaries),
	})
}

func GetAllDiaries(c *fiber.Ctx) error {
	var diaries []models.Diary

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	query := database.DB.Order("diary_date desc")

	if studentID := c.Query("student_id"); studentID != "" {
		query = query.Where("student_id = ?", studentID)
	}

	if isShared := c.Query("is_shared"); isShared != "" {
		query = query.Where("is_shared = ?", isShared)
	}

	if err := query.Offset(offset).Limit(limit).Find(&diaries).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to retrieve diaries",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Diaries retrieved successfully",
		"data":    diaries,
		"count":   len(diaries),
		"page":    page,
		"limit":   limit,
	})
}

func GetDiaryDateByStudentId(c *fiber.Ctx) error {
	studentId := c.Query("StudentID")

	if studentId == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "StudentID is required",
		})
	}

	var diaryDates []time.Time

	result := database.DB.Model(&models.Diary{}).
		Where("student_id = ?", studentId).
		Pluck("diary_date", &diaryDates)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	if len(diaryDates) == 0 {
		return c.JSON(fiber.Map{
			"message": "no diary dates found",
			"data":    []time.Time{},
		})
	}

	return c.JSON(fiber.Map{
		"message": "ok",
		"data":    diaryDates,
	})
}
