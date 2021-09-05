package model

import (
	"gorm.io/gorm"
)

type EventFile struct {
	gorm.Model
	EventID uint
	fileURL string
	mode    string
}

const (
	EventFileModePublic  = "public"
	EventFileModePrivate = "private"
	EventFileModePaid    = "paid"
)
