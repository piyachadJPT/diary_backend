package controllers

import (
	"gofiber-auth/database"
	"gofiber-auth/models"

	"github.com/gofiber/fiber/v2"
)

func CreateNewComment(c *fiber.Ctx) error {
	var comment models.Comment

	if err := c.BodyParser(&comment); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Cannot parse JSON",
			"comment": err.Error(),
		})
	}

	if comment.DiaryID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "DiaryID is required",
		})
	}

	if comment.AuthorID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "AuthorID is required",
		})
	}

	if comment.Content == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Content is required",
		})
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create diary",
			"details": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Created comment successfully",
		"data":    comment,
	})
}

func GetCommentByDiaryId(c *fiber.Ctx) error {
	diaryId := c.Query("DiaryID")
	if diaryId == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "DiaryID is required",
		})
	}

	var comments []models.Comment

	result := database.DB.
		Preload("Author").
		Where("diary_id = ?", diaryId).
		Find(&comments)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	if len(comments) == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "No comments found for this DiaryID",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Comments retrieved successfully",
		"data":    comments,
	})
}

func DeleteComment(c *fiber.Ctx) error {
	id := c.Params("id")
	var comment models.Comment

	if result := database.DB.First(&comment, id); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Comment not found",
		})
	}

	if err := database.DB.Delete(&comment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to delete comment",
			"details": err.Error(),
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"ok":      true,
		"message": "Deleted comment successfully",
	})
}
