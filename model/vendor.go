package model

import (
	"gorm.io/gorm"
)

type Vendor struct {
	gorm.Model
	Name            string
	Email           string
	Phone           string
	CompanyName     string
	Description     string
	Image           string
	Street1         string
	Street2         string
	City            string
	State           string
	Zip             string
	Website         string
	SocialFacebook  string
	SocialInstagram string
	SocialEtsy      string
	SocialWebsite   string
	SocialBohoSite  string

	Categories []VendorCategory `gorm:"ForeignKey:vendor_id"`

	Disabled                  bool
	Rejected                  bool
	StripeID                  string
	DefaultPaymentMethodLast4 string
	StripeConnectedAccountID  string
	ConnectionApplied         bool
}

//doing all the field mapping here so we can eventually do some validation, but also to merge in with fields that cant be updated
func (h *Vendor) Update(newData Vendor) {
	h.Name = newData.Name
	h.Phone = newData.Phone
	h.CompanyName = newData.CompanyName
	h.Description = newData.Description
	h.Image = newData.Image
	h.Street1 = newData.Street1
	h.Street2 = newData.Street2
	h.City = newData.City
	h.State = newData.State
	h.Zip = newData.Zip
	h.Website = newData.Website
	h.SocialFacebook = newData.SocialFacebook
	h.SocialInstagram = newData.SocialInstagram
	h.SocialEtsy = newData.SocialEtsy
	h.SocialWebsite = newData.SocialWebsite
	h.SocialBohoSite = newData.SocialBohoSite
}
