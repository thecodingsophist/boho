package model

import (
	"gorm.io/gorm"
)

type VenueCategory struct {
	gorm.Model
	VenueID    uint
	CategoryID uint     `gorm:"ForeignKey:id"`
	Category   Category `gorm:"ForeignKey:category_id"`
}
