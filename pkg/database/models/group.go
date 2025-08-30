package models

import (
	"time"
)

type Group struct {
	UID          uint      `gorm:"primaryKey;autoIncrement"`
	AddedDate    time.Time `gorm:"column:added_date"`
	ID           uint      `gorm:"unique;not null"`
	Years        string    `gorm:"size:255"`
	LocalName    string    `gorm:"size:255"`
	Name         string    `gorm:"size:255"`
	Specialty    string    `gorm:"size:255"`
	Level        string    `gorm:"size:255"`
	Course       int
	Abbreviation string `gorm:"size:255"`
	Faculty      string `gorm:"size:255"`
}

func (Group) TableName() string {
	return "groups"
}
