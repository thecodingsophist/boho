package model

import (
	"gorm.io/gorm"
)

type VendorPurchase struct {
	gorm.Model
	VendorID       uint
	EventID        uint
	EventProductID uint
}
