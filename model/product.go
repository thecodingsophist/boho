package model

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name            string
	Description     string
	BasePrice       uint
	StripeProductID string
}
