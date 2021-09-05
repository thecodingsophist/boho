package model

import (
	"gorm.io/gorm"
)

type VendorCart struct {
	gorm.Model
	VendorID       uint
	EventID        uint
	EventProductID uint
}
