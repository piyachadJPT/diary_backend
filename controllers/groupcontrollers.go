package controllers

import (
	"fmt"
	"gofiber-auth/database"
	"gofiber-auth/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetGroupsByAdvisor(c *fiber.Ctx) error {
	advisorIDStr := c.Query("advisor_id")

	if advisorIDStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Advisor ID is required",
		})
	}

	var groups []models.Group
	result := database.DB.Where("advisor_id = ?", advisorIDStr).Find(&groups)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Groups retrieved successfully",
		"data":    groups,
	})
}

func CreateGroup(c *fiber.Ctx) error {
	type Input struct {
		Name        string  `json:"name,omitempty"`
		Description *string `json:"description,omitempty"`
		AdvisorID   uint    `json:"advisor_id,omitempty"`
	}

	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid JSON body",
			"error":   err.Error(),
		})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Group name is required",
		})
	}

	if input.AdvisorID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Advisor ID is required",
		})
	}

	var advisor models.User
	if err := database.DB.First(&advisor, input.AdvisorID).Error; err != nil {
		fmt.Printf("Error finding advisor ID %d: %v\n", input.AdvisorID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": fmt.Sprintf("Advisor with ID %d not found", input.AdvisorID),
			"error":   err.Error(),
		})
	}

	if advisor.Role != "advisor" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": fmt.Sprintf("User with ID %d is not an advisor (role: %s)", input.AdvisorID, advisor.Role),
		})
	}

	group := models.Group{
		Name:        input.Name,
		Description: input.Description,
		AdvisorID:   input.AdvisorID,
	}

	if err := database.DB.Create(&group).Error; err != nil {
		fmt.Printf("Error creating group: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create group",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Group created successfully",
		"data":    group,
	})
}

func GetStudentsInGroup(c *fiber.Ctx) error {
	groupIDStr := c.Query("group_id")

	if groupIDStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Group ID is required",
		})
	}

	var studentGroups []models.StudentGroup
	result := database.DB.
		Preload("Student").
		Where("group_id = ?", groupIDStr).
		Find(&studentGroups)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	var students []fiber.Map
	for _, sg := range studentGroups {
		students = append(students, fiber.Map{
			"id":              sg.Student.ID,
			"name":            sg.Student.Name,
			"email":           sg.Student.Email,
			"image":           sg.Student.Image,
			"created_at":      sg.Student.CreatedAt,
			"joined_group_at": sg.CreatedAt,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Students retrieved successfully",
		"data":    students,
	})
}

func GetStudentsWithoutGroup(c *fiber.Ctx) error {
	advisorIDStr := c.Query("advisor_id")

	if advisorIDStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Advisor ID is required",
		})
	}

	var studentAdvisors []models.StudentAdvisor
	result := database.DB.
		Preload("Student").
		Where("advisor_id = ?", advisorIDStr).
		Find(&studentAdvisors)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	var studentIdsInGroups []uint
	database.DB.Model(&models.StudentGroup{}).
		Joins("JOIN groups ON student_groups.group_id = groups.id").
		Where("groups.advisor_id = ?", advisorIDStr).
		Pluck("student_groups.student_id", &studentIdsInGroups)

	var studentsWithoutGroup []fiber.Map
	for _, sa := range studentAdvisors {
		isInGroup := false
		for _, id := range studentIdsInGroups {
			if sa.StudentID == id {
				isInGroup = true
				break
			}
		}
		if !isInGroup {
			studentsWithoutGroup = append(studentsWithoutGroup, fiber.Map{
				"id":         sa.Student.ID,
				"name":       sa.Student.Name,
				"email":      sa.Student.Email,
				"image":      sa.Student.Image,
				"created_at": sa.Student.CreatedAt,
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Students without group retrieved successfully",
		"data":    studentsWithoutGroup,
	})
}

func AddStudentToGroup(c *fiber.Ctx) error {
	type Input struct {
		StudentID uint `json:"student_id"`
		GroupID   uint `json:"group_id"`
	}

	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid JSON body",
			"error":   err.Error(),
		})
	}

	if input.StudentID == 0 || input.GroupID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Student ID and Group ID are required and must be greater than 0",
		})
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to begin transaction",
		})
	}

	var student models.User
	if err := tx.First(&student, input.StudentID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Student not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error checking student",
			"error":   err.Error(),
		})
	}

	var group models.Group
	if err := tx.First(&group, input.GroupID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Group not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error checking group",
			"error":   err.Error(),
		})
	}

	var existing models.StudentGroup
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("student_id = ? AND group_id = ?", input.StudentID, input.GroupID).
		First(&existing).Error

	if err == nil {
		tx.Rollback()
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Student is already in this group",
		})
	} else if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error when checking existing relationship",
			"error":   err.Error(),
		})
	}

	studentGroup := models.StudentGroup{
		StudentID: input.StudentID,
		GroupID:   input.GroupID,
	}

	if err := tx.Create(&studentGroup).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to add student to group",
			"error":   err.Error(),
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to commit transaction",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Student added to group successfully",
		"data":    studentGroup,
	})
}

func RemoveStudentFromGroup(c *fiber.Ctx) error {
	studentIDStr := c.Query("student_id")
	groupIDStr := c.Query("group_id")

	if studentIDStr == "" || groupIDStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Student ID and Group ID are required",
		})
	}

	result := database.DB.Where("student_id = ? AND group_id = ?", studentIDStr, groupIDStr).
		Delete(&models.StudentGroup{})

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "Student-group relationship not found",
		})
	}

	return c.JSON(fiber.Map{
		"ok":      true,
		"message": "Student removed from group successfully",
	})
}

func DeleteGroup(c *fiber.Ctx) error {
	idParam := c.Query("id")
	if idParam == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Group ID is required",
		})
	}

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	tx := database.DB.Begin()

	if err := tx.Where("group_id = ?", id).Delete(&models.StudentGroup{}).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete group relationships",
		})
	}

	result := tx.Delete(&models.Group{}, id)
	if result.Error != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{
			"error": "Group not found",
		})
	}

	tx.Commit()

	return c.JSON(fiber.Map{
		"ok":      true,
		"message": "Group deleted successfully",
	})
}
