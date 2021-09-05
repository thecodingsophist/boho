package model

import (
	"gorm.io/gorm"
)

type VendorCategory struct {
	gorm.Model
	VendorID   uint
	CategoryID uint     `gorm:"ForeignKey:id"`
	Category   Category `gorm:"ForeignKey:category_id"`
}
