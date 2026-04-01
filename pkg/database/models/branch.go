package models

type Branch struct {
	ID uint `gorm:"primaryKey"`

	Code string `gorm:"size:32;not null;uniqueIndex"` // e.g. "main", "north", "spb"
	Name string `gorm:"size:128;not null"`            // display name

	// Location
	Country string `gorm:"size:2;not null"` // ISO 3166-1 alpha-2, e.g. "LV"
	City    string `gorm:"size:64;not null"`
	Address string `gorm:"size:256"`

	// Timezone (IANA), e.g. "Europe/Moscow"
	Timezone string `gorm:"size:64;not null;default:Europe/Moscow"`

	// has many Groups
	Groups []Group `gorm:"foreignKey:BranchID;references:ID"`
	// has many Faculties
	Faculties []Faculty `gorm:"foreignKey:BranchID;references:ID"`
}
