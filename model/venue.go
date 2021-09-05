package model

import (
	"gorm.io/gorm"
)

type Venue struct {
	gorm.Model
	Name             string
	Street1          string
	Street2          string
	City             string
	State            string
	Zip              string
	Email            string
	Phone            string
	Description      string
	Hours            string
	FAQ              string
	Events           []*Event `gorm:"ForeignKey:venue_id"`
	AllowFood        bool
	AllowFoodTruck   bool
	VendorSpaces     uint
	FoodTruckSpaces  uint
	PriceSingleSpace uint
	PriceDoubleSpace uint
	PriceFoodSpace   uint
	PriceTruckSpace  uint
	EventGuide       string
	BoothLayout      string
	Categories       []VenueCategory `gorm:"ForeignKey:venue_id"`
}
