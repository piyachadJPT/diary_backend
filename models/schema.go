package models

import (
	"time"

	"gorm.io/datatypes"
)

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      *string   `gorm:"size:255"`
	Email     string    `gorm:"size:255;unique;not null"`
	Password  *string   `gorm:"size:255"`
	Approved  bool      `gorm:"not null;default:true"`
	Image     *string   `gorm:"type:longtext"`
	Role      string    `gorm:"type:enum('student','advisor');not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type StudentAdvisor struct {
	ID        uint `gorm:"primaryKey;autoIncrement"`
	AdvisorID uint `gorm:"not null;index"`
	StudentID uint `gorm:"not null;index"`

	Advisor User `gorm:"foreignKey:AdvisorID;references:ID;constraint:OnDelete:CASCADE"`
	Student User `gorm:"foreignKey:StudentID;references:ID;constraint:OnDelete:CASCADE"`
}

type Diary struct {
	ID           uint           `gorm:"primaryKey;autoIncrement"`
	StudentID    uint           `gorm:"not null;index"`
	ContentHTML  string         `gorm:"type:text"`
	ContentDelta datatypes.JSON `gorm:"type:json"`
	IsShared     string         `gorm:"default:everyone"`
	AllowComment bool           `gorm:"default:true"`
	Status       string         `gorm:"default:neutral"`
	DiaryDate    time.Time      `gorm:"type:date;not null"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`

	Student     User         `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	Attachments []Attachment `gorm:"foreignKey:DiaryID"`
}

type Attachment struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	DiaryID   uint      `gorm:"not null;index"`
	FileURL   string    `gorm:"type:text;not null"`
	FileName  string    `gorm:"size:255;not null"`
	FileType  string    `gorm:"size:100"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	Diary Diary `gorm:"-:all" json:"-"`
}

type Comment struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	DiaryID   uint      `gorm:"not null;index"`
	AuthorID  uint      `gorm:"not null;index"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	Diary  Diary `gorm:"foreignKey:DiaryID;constraint:OnDelete:CASCADE"`
	Author User  `gorm:"foreignKey:AuthorID;constraint:OnDelete:CASCADE"`
}

type Notification struct {
	ID        uint           `gorm:"primaryKey;autoIncrement"`
	UserID    uint           `gorm:"not null;index"`
	DiaryID   *uint          `gorm:"index"`
	Type      string         `gorm:"size:50;not null"`
	Title     string         `gorm:"size:255;not null"`
	Message   string         `gorm:"type:text;not null"`
	Data      datatypes.JSON `gorm:"type:json"`
	IsRead    bool           `gorm:"default:false"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`

	User  User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Diary *Diary `gorm:"foreignKey:DiaryID;constraint:OnDelete:CASCADE"`
}
