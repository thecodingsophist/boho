package model

import (
	"gorm.io/gorm"
)

type ApplicationInventory struct {
	gorm.Model
	ProductID     uint `gorm:"ForeignKey:id"`
	Quantity      uint
	BasePrice     uint
	ApplicationID uint        `gorm:"ForeignKey:id"`
	Product       Product     `gorm:"ForeignKey:product_id"`
	Application   Application `gorm:"ForeignKey:application_id"`
}
