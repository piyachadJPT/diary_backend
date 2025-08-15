package controllers

import (
	"gofiber-auth/database"
	"gofiber-auth/models"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginData struct {
	Email string  `json:"email"`
	Name  *string `json:"name"`
	Image *string `json:"image"`
}

var jwtSecret = []byte(os.Getenv("CORS_ALLOW_SECRET"))

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

func LoginHandler(c *fiber.Ctx) error {
	type LoginInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	if input.Email == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	var user models.User
	if err := database.DB.Where("email = ? AND approved = ?", input.Email, true).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "บัญชีของคุณยังไม่ได้รับการอนุมัติ",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(input.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Incorrect password",
		})
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role, // เพิ่ม role
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not generate token",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"token":   signedToken,
		"role":    user.Role,
	})
}
