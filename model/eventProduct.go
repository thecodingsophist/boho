package model

import (
	"gorm.io/gorm"
)

type EventProduct struct {
	gorm.Model
	ProductID uint `gorm:"ForeignKey:id"`
	Quantity  uint
	BasePrice uint
	EventID   uint    `gorm:"ForeignKey:id"`
	Product   Product `gorm:"ForeignKey:product_id"`
	Event     Event   `gorm:"ForeignKey:event_id"`
}