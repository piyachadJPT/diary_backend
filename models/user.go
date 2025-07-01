package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Name      string    `gorm:"not null"`
	Role      string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}
