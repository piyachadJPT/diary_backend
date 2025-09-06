package controllers

import (
	"gofiber-auth/database"
	"gofiber-auth/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `json:"ID"`
	Name      string    `json:"Name"`
	Email     string    `json:"Email"`
	Password  string    `json:"-"`
	Approved  bool      `json:"Approved"`
	Image     string    `json:"Image"`
	Role      string    `json:"Role"`
	CreatedAt time.Time `json:"CreatedAt"`
}

func GetAllStudents(c *fiber.Ctx) error {
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

func GetStudentById(c *fiber.Ctx) error {
	id := c.Query("id")
	var student models.User

	result := database.DB.Where("id = ?", id).First(&student)
	if result.Error != nil {
		c.Status(404)
		return c.SendString(`student not found`)
	}

	return c.JSON(student)
}

func CreateStudent(c *fiber.Ctx) error {
	type Input struct {
		Email     string `json:"email"`
		AdvisorID uint   `json:"advisor_id"`
	}

	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid JSON body",
		})
	}

	if input.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Email is required",
		})
	}

	var advisor models.User
	if err := database.DB.First(&advisor, input.AdvisorID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Advisor not found",
		})
	}
	if advisor.Role != "advisor" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Provided user is not an advisor",
		})
	}

	var student models.User
	err := database.DB.Where("email = ?", input.Email).First(&student).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			student = models.User{
				Email: input.Email,
				Role:  "student",
			}
			if err := database.DB.Create(&student).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "Failed to create student user",
				})
			}
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error when checking for existing email",
			})
		}
	}

	var existingRelation models.StudentAdvisor
	if err := database.DB.Where("advisor_id = ? AND student_id = ?", advisor.ID, student.ID).
		First(&existingRelation).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "คุณได้เพิ่มนิสิตคนนี้ให้เป็นที่ปรึกษาแล้ว",
		})
	} else if err != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error when checking for existing relationship",
		})
	}

	studentAdvisor := models.StudentAdvisor{
		AdvisorID: advisor.ID,
		StudentID: student.ID,
	}
	if err := database.DB.Create(&studentAdvisor).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create student-advisor relation",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Student and advisor relationship created successfully",
		"user_id": student.ID,
	})
}
