package models

type Speciality struct {
	ID uint `gorm:"primaryKey"`

	BranchID uint   `gorm:"not null;index"`
	Branch   Branch `gorm:"foreignKey:BranchID;references:ID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT;"`

	Name  string `gorm:"size:256;not null"`
	Level string `gorm:"size:256"`

	Groups []Group `gorm:"foreignKey:SpecialityID;references:ID"`
}
