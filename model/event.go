package model

import (
	"gorm.io/gorm"
	"time"
)

type Event struct {
	gorm.Model
	Name             string
	Description      string
	Hours            string
	StartDate        time.Time
	EndDate          time.Time
	ImageURL         string
	FAQ              string
	VenueID          uint `gorm:"ForeignKey:id"`
	EventStatus      string
	Venue            Venue `gorm:"ForeignKey:venue_id"`
	VendorSpaces     uint
	FoodTruckSpaces  uint
	AllowFood        bool
	AllowFoodTruck   bool
	PriceSingleSpace uint
	PriceDoubleSpace uint
	PriceFoodSpace   uint
	PriceTruckSpace  uint
	EventGuide       string
	BoothLayout      string
	Categories       []EventCategory `gorm:"ForeignKey:event_id"`
	Participants     []Application   `gorm:"ForeignKey:event_id"`
	Files            []EventFile     `gorm:"ForeignKey:event_id"`
}

const (
	EventStatusUnpublished = "Unpublished"
	EventStatusOpen        = "open"
	EventStatusClosed      = "closed"
	EventStatusCancelled   = "cancelled"
)
