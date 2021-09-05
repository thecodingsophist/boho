package model

import (
	"gorm.io/gorm"
)

type EventCategory struct {
	gorm.Model
	EventID    uint
	CategoryID uint     `gorm:"ForeignKey:id"`
	Category   Category `gorm:"ForeignKey:category_id"`
}
