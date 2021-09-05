package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UID       string //google firebase auth identifier
	Email     string
	IsAdmin   bool
	Onboarded bool
	Disabled  bool
}
