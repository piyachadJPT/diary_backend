package controllers

import (
	"fmt"
	"gofiber-auth/database"
	"gofiber-auth/models"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User
	result := database.DB.First(&user, id)
	if result.Error != nil {
		c.Status(404)
		return c.SendString("User not found")
	}
	return c.JSON(user)
}

func GetUserByEmail(c *fiber.Ctx) error {
	email := c.Query("email")
	var user models.User

	result := database.DB.Where("email = ?", email).First(&user)

	if result.Error != nil {
		c.Status(404)
		return c.SendString("User not found")
	}

	return c.JSON(user)
}

func GetAllUser(c *fiber.Ctx) error {
	var user models.User

	result := database.DB.Find(&user)
	if result.Error != nil {
		c.Status(404)
		return c.SendString("User not found")
	}

	return c.JSON(user)
}

func CreateUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		c.Status(400)
		return c.SendString("Cannot parse JSON")
	}

	database.DB.Create(&user)

	return c.SendString("User created successfully")
}

// func RegisterHandler(c *fiber.Ctx) error {
// 	type RegisterInput struct {
// 		Name     string `json:"name"`
// 		Email    string `json:"email"`
// 		Password string `json:"password"`
// 	}

// 	var input RegisterInput
// 	if err := c.BodyParser(&input); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid input",
// 		})
// 	}

// 	if input.Email == "" || input.Password == "" {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Email and password are required",
// 		})
// 	}

// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Could not hash password",
// 		})
// 	}

// 	user := models.User{
// 		Name:     &input.Name,
// 		Email:    input.Email,
// 		Password: func() *string { s := string(hashedPassword); return &s }(),
// 	}

// 	if err := database.DB.Create(&user).Error; err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Email already exists",
// 		})
// 	}

// 	if err := database.DB.Model(&user).Update("Approved", false).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Could not set Approved to false",
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"message": "User registered successfully",
// 	})
// }

func RegisterHandler(c *fiber.Ctx) error {
	type RegisterInput struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var input RegisterInput
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not hash password",
		})
	}

	user := models.User{
		Name:     &input.Name,
		Email:    input.Email,
		Password: func() *string { s := string(hashedPassword); return &s }(),
		Role:     "student", // ตั้ง role เป็น student
	}

	// สร้าง user
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email already exists",
		})
	}

	// อัปเดต Approved เป็น false หลังสร้าง
	if err := database.DB.Model(&user).Update("Approved", false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not set Approved to false",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User registered successfully",
	})
}

func GetAllUserNotconfirmed(c *fiber.Ctx) error {
	var users []models.User

	result := database.DB.Where("approved = ?", 0).Find(&users)
	if result.Error != nil {
		c.Status(500)
		return c.SendString("Database error")
	}

	if len(users) == 0 {
		c.Status(404)
		return c.SendString("No unconfirmed users found")
	}

	return c.JSON(users)
}

func PatchApproved(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User

	if result := database.DB.First(&user, id); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	var updateApproved map[string]interface{}
	if err := c.BodyParser(&updateApproved); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Cannot parse JSON",
			"details": err.Error(),
		})
	}

	if err := database.DB.Model(&user).Updates(updateApproved).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to update approved",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Patched approved successfully",
	})
}

func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing token"})
	}

	// ตัด "Bearer " ออก
	tokenStr := strings.Replace(authHeader, "Bearer ", "", 1)

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
	}

	// ดึง user_id จาก claims
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user_id in token"})
	}
	userID := uint(userIDFloat)

	// query user จาก database
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// เก็บ user ใน context เพื่อ handler ต่อไปใช้
	c.Locals("user", &user)
	return c.Next()
}

func GetProfileHandler(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(fiber.Map{
		"id":       user.ID,
		"name":     user.Name,
		"email":    user.Email,
		"role":     user.Role,
		"approved": user.Approved,
		"image":    user.Image,
	})
}
