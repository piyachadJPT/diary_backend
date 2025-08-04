package controllers

import (
	"errors"
	"fmt"
	"gofiber-auth/database"
	"gofiber-auth/models"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func UploadAttachment(c *fiber.Ctx) error {
	diaryID, err := strconv.Atoi(c.FormValue("diary_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "diary_id ไม่ถูกต้อง",
		})
	}

	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ไม่สามารถอ่านไฟล์ได้",
		})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "ไม่มีไฟล์ที่อัปโหลด",
		})
	}

	savePath := fmt.Sprintf("upload/diary/")
	if err := os.MkdirAll(savePath, os.ModePerm); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถสร้างโฟลเดอร์ได้",
		})
	}

	var uploadedFiles []fiber.Map
	var attachments []models.Attachment

	for _, file := range files {
		allowedTypes := []string{
			"image/jpeg", "image/png", "image/gif", "image/webp",
			"video/mp4", "video/webm", "video/quicktime",
			"audio/mp3", "audio/wav", "audio/mpeg",
			"application/pdf", "application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"text/plain", "application/zip", "application/x-rar-compressed",
		}

		var contentType string
		if file.Header != nil && file.Header["Content-Type"] != nil && len(file.Header["Content-Type"]) > 0 {
			contentType = file.Header["Content-Type"][0]
		} else {
			ext := strings.ToLower(filepath.Ext(file.Filename))
			if ext == ".pdf" {
				contentType = "application/pdf"
			} else {
				contentType = ""
			}
		}

		if contentType != "application/pdf" {
			return errors.New("only PDF files are allowed")
		}

		isAllowed := false
		for _, allowedType := range allowedTypes {
			if contentType == allowedType {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return c.Status(400).JSON(fiber.Map{
				"error": fmt.Sprintf("ไฟล์ %s ไม่ได้รับอนุญาต (ประเภท: %s)", file.Filename, contentType),
			})
		}

		newFileName := uuid.New().String() + "_" + file.Filename
		fullPath := filepath.Join(savePath, newFileName)

		if err := c.SaveFile(file, fullPath); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("ไม่สามารถบันทึกไฟล์ %s ได้: %v", file.Filename, err),
			})
		}

		attachment := models.Attachment{
			DiaryID:  uint(diaryID),
			FileURL:  fullPath,
			FileName: file.Filename,
			FileType: contentType,
		}

		attachments = append(attachments, attachment)
		uploadedFiles = append(uploadedFiles, fiber.Map{
			"original_name": file.Filename,
			"saved_name":    newFileName,
			"file_path":     fullPath,
			"file_size":     file.Size,
			"content_type":  contentType,
		})
	}

	if err := database.DB.Create(&attachments).Error; err != nil {
		for _, file := range uploadedFiles {
			os.Remove(file["file_path"].(string))
		}
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("ไม่สามารถบันทึกข้อมูลลงฐานข้อมูลได้: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message":        "อัปโหลดสำเร็จ",
		"uploaded_files": uploadedFiles,
		"total_files":    len(uploadedFiles),
	})
}

func DeleteAttachment(c *fiber.Ctx) error {
	attachmentID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID ไม่ถูกต้อง",
		})
	}

	var attachment models.Attachment
	if err := database.DB.First(&attachment, attachmentID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "ไม่พบไฟล์แนบ",
		})
	}

	if err := os.Remove(attachment.FileURL); err != nil {
		fmt.Printf("Warning: ไม่สามารถลบไฟล์ %s ได้: %v\n", attachment.FileURL, err)
	}

	if err := database.DB.Delete(&attachment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถลบข้อมูลจากฐานข้อมูลได้",
		})
	}

	return c.JSON(fiber.Map{
		"message": "ลบไฟล์สำเร็จ",
	})
}

func GetAttachmentsByDiaryId(c *fiber.Ctx) error {
	idStr := c.Query("ID")
	if idStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id is required"})
	}

	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	id := uint(id64)

	var attachments []models.Attachment
	result := database.DB.
		Where("ID = ?", id).
		Find(&attachments)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": result.Error.Error()})
	}

	if len(attachments) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "No attachments found for this Diary ID"})
	}

	return c.JSON(fiber.Map{
		"message":     "Attachments retrieved successfully",
		"attachments": attachments,
	})
}
