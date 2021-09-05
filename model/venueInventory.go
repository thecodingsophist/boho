package model

import (
	"gorm.io/gorm"
)

type VenueInventory struct {
	gorm.Model
	EventProductID uint         `gorm:"ForeignKey:id"`
	Product        EventProduct `gorm:"ForeignKey:event_product_id"`
	Quantity       uint
	BasePrice      uint
	VenueID        uint  `gorm:"ForeignKey:id"`
	Venue          Venue `gorm:"ForeignKey:venue_id"`
}
