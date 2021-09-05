package service

import (
	"errors"
	"gorm.io/gorm"

	"hawker/model"
)

type User struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *User {
	return &User{db: db}
}

//does a fresh pull of things for a vendor
func (h *User) GetVendor(id uint) (model.Vendor, error) {
	var ve model.Vendor
	h.db.Preload("Categories").Preload("Categories.Category").First(&ve, id)
	if ve.ID < 1 {
		return ve, errors.New("no such vendor")
	}
	return ve, nil
}

func (h *User) GetVendorByEmail(email string) (model.Vendor, error) {
	var ve model.Vendor
	h.db.Preload("Categories").Preload("Categories.Category").First(&ve, "email = ?", email)
	if ve.ID < 1 {
		return ve, errors.New("no such vendor")
	}
	return ve, nil
}

func (h *User) ListVendors() ([]model.Vendor, error) {
	var ve []model.Vendor
	h.db.Preload("Categories").Preload("Categories.Category").Find(&ve)
	return ve, nil
}
