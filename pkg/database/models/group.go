package models

import (
	"database/sql"
)

type Group struct {
	ID uint `gorm:"primaryKey"`

	BranchID uint   `gorm:"not null;index"`
	Branch   Branch `gorm:"foreignKey:BranchID;references:ID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT;"`

	LocalName sql.NullString
	Name      string `gorm:"not null"`
	Course    uint8  `gorm:"not null"`

	SpecialityID uint       `gorm:"not null;index"`
	Speciality   Speciality `gorm:"foreignKey:SpecialityID;references:ID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT;"`

	FacultyID uint    `gorm:"not null;index"`
	Faculty   Faculty `gorm:"foreignKey:FacultyID;references:ID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT;"`

	Couples []Couple `gorm:"foreignKey:GroupID;references:ID"`
}
