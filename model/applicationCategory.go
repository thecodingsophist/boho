package model

import (
	"gorm.io/gorm"
)

type ApplicationCategory struct {
	gorm.Model
	ApplicationID uint
	CategoryID    uint
}
