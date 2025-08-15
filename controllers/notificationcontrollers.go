package controllers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gofiber-auth/database"
	"gofiber-auth/models"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AdvisorConnection struct {
	Ch       chan models.Notification
	LastSeen time.Time
}

var (
	advisorChannels = make(map[uint]*AdvisorConnection)
	channelMutex    sync.RWMutex
)

func cleanupInactiveConnections() {
	channelMutex.Lock()
	defer channelMutex.Unlock()

	now := time.Now()
	for advisorID, conn := range advisorChannels {
		if now.Sub(conn.LastSeen) > time.Minute*5 {
			close(conn.Ch)
			delete(advisorChannels, advisorID)
			fmt.Printf("Cleaned up inactive connection for advisor %d\n", advisorID)
		}
	}
}

func init() {
	go func() {
		ticker := time.NewTicker(time.Minute * 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cleanupInactiveConnections()
			}
		}
	}()
}

func AdvisorSSE(c *fiber.Ctx) error {
	advisorID, err := strconv.ParseUint(c.Query("advisor_id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid advisor_id"})
	}

	var advisorExists bool
	if err := database.DB.Model(&models.User{}).
		Select("count(*) > 0").
		Where("id = ? AND role = ?", advisorID, "advisor").
		Find(&advisorExists).Error; err != nil || !advisorExists {
		return c.Status(404).JSON(fiber.Map{"error": "advisor not found"})
	}

	c.Set("Access-Control-Allow-Origin", os.Getenv("CORS_ALLOW_ORIGINS"))
	c.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Cache-Control")
	c.Set("Access-Control-Allow-Credentials", "true")
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	channelMutex.Lock()
	if conn, exists := advisorChannels[uint(advisorID)]; exists {
		close(conn.Ch)
		delete(advisorChannels, uint(advisorID))
		fmt.Printf("Closed existing connection for advisor %d\n", advisorID)
	}

	conn := &AdvisorConnection{
		Ch:       make(chan models.Notification, 50),
		LastSeen: time.Now(),
	}
	advisorChannels[uint(advisorID)] = conn
	channelMutex.Unlock()

	defer func() {
		channelMutex.Lock()
		if existingConn, stillExists := advisorChannels[uint(advisorID)]; stillExists && existingConn == conn {
			close(conn.Ch)
			delete(advisorChannels, uint(advisorID))
			fmt.Printf("Connection closed for advisor %d\n", advisorID)
		}
		channelMutex.Unlock()
	}()

	connectionMsg := map[string]interface{}{
		"type":    "connected",
		"message": "SSE connection established",
		"time":    time.Now().Format(time.RFC3339),
	}
	if connData, err := json.Marshal(connectionMsg); err == nil {
		c.WriteString(fmt.Sprintf("data: %s\n\n", string(connData)))
	}

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic in SSE for advisor %d: %v\n", advisorID, r)
			}
		}()

		if conn == nil || conn.Ch == nil {
			fmt.Printf("Connection or channel is nil for advisor %d\n", advisorID)
			return
		}

		heartbeatTicker := time.NewTicker(time.Second * 30)
		defer heartbeatTicker.Stop()

		for {
			select {
			case notif, ok := <-conn.Ch:
				if !ok || conn == nil {
					fmt.Printf("Channel closed or connection nil for advisor %d\n", advisorID)
					return
				}

				channelMutex.Lock()
				if existingConn, exists := advisorChannels[uint(advisorID)]; exists && existingConn == conn {
					conn.LastSeen = time.Now()
				}
				channelMutex.Unlock()

				data, err := json.Marshal(notif)
				if err != nil {
					fmt.Printf("Error marshaling notification: %v\n", err)
					continue
				}

				fmt.Fprintf(w, "data: %s\n\n", string(data))
				if err := w.Flush(); err != nil {
					fmt.Printf("Error flushing data for advisor %d: %v\n", advisorID, err)
					return
				}

			case <-heartbeatTicker.C:
				channelMutex.Lock()
				if existingConn, exists := advisorChannels[uint(advisorID)]; exists && existingConn == conn {
					conn.LastSeen = time.Now()
				}
				channelMutex.Unlock()

				heartbeat := map[string]interface{}{
					"type": "heartbeat",
					"time": time.Now().Format(time.RFC3339),
				}
				if heartbeatData, err := json.Marshal(heartbeat); err == nil {
					fmt.Fprintf(w, "data: %s\n\n", string(heartbeatData))
				} else {
					fmt.Fprintf(w, ": heartbeat\n\n")
				}

				if err := w.Flush(); err != nil {
					fmt.Printf("Error sending heartbeat for advisor %d: %v\n", advisorID, err)
					return
				}

			case <-c.Context().Done():
				fmt.Printf("Context done for advisor %d\n", advisorID)
				return
			}
		}
	})

	return nil
}

func SendNotificationToAdvisor(advisorID uint, notification models.Notification) {
	channelMutex.RLock()
	conn, exists := advisorChannels[advisorID]
	channelMutex.RUnlock()

	if !exists {
		fmt.Printf("No active connection for advisor %d\n", advisorID)
		return
	}

	select {
	case conn.Ch <- notification:
		fmt.Printf("Notification sent to advisor %d\n", advisorID)
	default:
		fmt.Printf("Channel full for advisor %d, notification dropped\n", advisorID)
	}
}

func GetUnreadNotifications(c *fiber.Ctx) error {
	advisorID := c.Query("advisor_id")
	if advisorID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "advisor_id is required",
		})
	}

	var notifications []models.Notification
	result := database.DB.
		Preload("Diary").
		Preload("Diary.Student").
		Preload("User").
		Where("user_id = ? AND is_read = ?", advisorID, false).
		Order("created_at DESC").
		Find(&notifications)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Unread notifications retrieved successfully",
		"data":    notifications,
		"count":   len(notifications),
	})
}

func GetAllNotifications(c *fiber.Ctx) error {
	advisorID := c.Query("advisor_id")
	if advisorID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "advisor_id is required",
		})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	if limit > 100 {
		limit = 100
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	var notifications []models.Notification
	result := database.DB.
		Preload("Diary").
		Preload("Diary.Student").
		Preload("User").
		Where("user_id = ?", advisorID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.JSON(fiber.Map{
				"message": "No notifications found",
				"data":    []models.Notification{},
				"count":   0,
				"total":   0,
				"page":    page,
				"limit":   limit,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	var total int64
	if err := database.DB.Model(&models.Notification{}).Where("user_id = ?", advisorID).Count(&total).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to count notifications",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Notifications retrieved successfully",
		"data":    notifications,
		"count":   len(notifications),
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

func MarkNotificationAsRead(c *fiber.Ctx) error {
	notificationID := c.Params("id")
	if notificationID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "notification ID is required",
		})
	}

	var notification models.Notification
	if result := database.DB.Preload("User").First(&notification, notificationID); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"error": "Notification not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	// ถ้าอ่านแล้วก็ไม่ต้องอัพเดท
	if notification.IsRead {
		return c.JSON(fiber.Map{
			"message": "Notification already marked as read",
			"data":    notification,
		})
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	notification.IsRead = true

	if err := tx.Save(&notification).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to update notification",
			"details": err.Error(),
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	readEvent := models.Notification{
		ID:     notification.ID,
		Type:   "notification_read",
		IsRead: true,
	}
	SendNotificationToAdvisor(notification.UserID, readEvent)

	return c.JSON(fiber.Map{
		"message": "Notification marked as read",
		"data":    notification,
	})
}

func MarkAllNotificationsAsRead(c *fiber.Ctx) error {
	advisorID := c.Query("advisor_id")
	if advisorID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "advisor_id is required",
		})
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", advisorID, false).
		Update("is_read", true)

	if result.Error != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Marked %d notifications as read", result.RowsAffected),
		"count":   result.RowsAffected,
	})
}

func DeleteNotification(c *fiber.Ctx) error {
	notificationID := c.Params("id")
	if notificationID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "notification ID is required",
		})
	}

	var notification models.Notification
	if result := database.DB.First(&notification, notificationID); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"error": "Notification not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	if err := database.DB.Delete(&notification).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to delete notification",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Notification deleted successfully",
	})
}

func GetNotificationCount(c *fiber.Ctx) error {
	advisorID := c.Query("advisor_id")
	if advisorID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "advisor_id is required",
		})
	}

	var unreadCount int64
	result := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", advisorID, false).
		Count(&unreadCount)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}

	var totalCount int64
	database.DB.Model(&models.Notification{}).
		Where("user_id = ?", advisorID).
		Count(&totalCount)

	return c.JSON(fiber.Map{
		"message":      "Notification count retrieved successfully",
		"unread_count": unreadCount,
		"total_count":  totalCount,
	})
}

func GetActiveConnections(c *fiber.Ctx) error {
	channelMutex.RLock()
	defer channelMutex.RUnlock()

	connections := make(map[uint]interface{})
	for advisorID, conn := range advisorChannels {
		connections[advisorID] = map[string]interface{}{
			"last_seen": conn.LastSeen.Format(time.RFC3339),
			"active":    time.Since(conn.LastSeen) < time.Minute*5,
		}
	}

	return c.JSON(fiber.Map{
		"message":     "Active connections retrieved",
		"connections": connections,
		"count":       len(connections),
	})
}
