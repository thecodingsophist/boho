package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hanzoai/gochimp3"
	"gorm.io/gorm"
	"net/http"

	"hawker/model"
)

type Vendor struct {
	db *gorm.DB
}

func NewVendorController(d *gorm.DB) *Vendor {
	return &Vendor{
		db: d,
	}
}

//combining the shared work of create and edit as much as possible
func (h *Vendor) Edit(c *gin.Context) {
	var data vendor
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	var newVendor bool
	if data.ID < 1 {
		newVendor = true
	}
	//TODO: data validation
	vend := vendorToModel(data)

	//Security check
	admin, ok := c.Get("isAdmin")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "admin key failure",
		})
		return
	}
	isAdmin, _ := admin.(bool)
	email, ok := c.Get("email")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "admin key failure",
		})
		return
	}
	if !isAdmin && email.(string) != vend.Email {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "only admins may edit other vendor profiles",
		})
		return
	}

	//trying to fix the update process so it doesnt lose info
	if !newVendor {
		var oldRecord model.Vendor
		h.db.Find(&oldRecord, data.ID)
		oldRecord.Update(vend)
		vend = oldRecord
	}

	//hack it right into the vendor save for now...
	if data.MailingList {
		mlC := gochimp3.New("bff2ad08f3bcc1e12783c897d5dd11d1-us13")
		list, err := mlC.GetList("bd0d932998", nil)
		if err != nil {
			fmt.Println(err)
		}
		req := &gochimp3.MemberRequest{
			EmailAddress: data.Email,
			Status:       "subscribed",
		}
		_, err = list.CreateMember(req)
		if err != nil {
			fmt.Println(err)
		}
	}

	h.db.Save(&vend)

	var vcat []model.VendorCategory
	h.db.Delete(&vcat, "vendor_id = ?", vend.ID)

	for _, cat := range data.Categories {
		ca := model.VendorCategory{
			VendorID:   vend.ID,
			CategoryID: cat.ID,
		}
		vcat = append(vcat, ca)
	}

	h.db.Save(&vcat)

	//onboard the specified user
	var u model.User
	h.db.First(&u, "email = ?", data.Email)
	u.Onboarded = true
	if newVendor {
		u.Disabled = true
	}
	h.db.Save(&u)

	out, err := h.getVendor(vend.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no such vendor found",
		})
		return
	}

	c.JSON(http.StatusOK, vendorTranslation(out))
}

//does a fresh pull of things for a vendor
func (h *Vendor) getVendor(id uint) (model.Vendor, error) {
	var ve model.Vendor
	h.db.Preload("Categories").Preload("Categories.Category").First(&ve, id)
	if ve.ID < 1 {
		return ve, errors.New("no such vendor")
	}
	return ve, nil
}

func vendorTranslation(v model.Vendor) vendor {
	vr := vendor{
		ID:                        v.ID,
		Name:                      v.Name,
		Email:                     v.Email,
		Phone:                     v.Phone,
		CompanyName:               v.CompanyName,
		Description:               v.Description,
		Image:                     v.Image,
		Street1:                   v.Street1,
		Street2:                   v.Street2,
		City:                      v.City,
		State:                     v.State,
		Zip:                       v.Zip,
		Website:                   v.Website,
		Disabled:                  v.Disabled,
		SocialFacebook:            v.SocialFacebook,
		SocialInstagram:           v.SocialInstagram,
		SocialEtsy:                v.SocialEtsy,
		SocialWebsite:             v.SocialWebsite,
		SocialBohoSite:            v.SocialBohoSite,
		DefaultPaymentMethodLast4: v.DefaultPaymentMethodLast4,
		StripeID:                  v.StripeID,
		Rejected:                  v.Rejected,
		StripeConnectedAccountID:  v.StripeConnectedAccountID,
		ConnectionApplied:         v.ConnectionApplied,
	}
	for _, cat := range v.Categories {
		ca := category{
			ID:          cat.CategoryID,
			Name:        cat.Category.Name,
			Description: cat.Category.Description,
		}
		vr.Categories = append(vr.Categories, ca)
	}
	return vr
}

func vendorToModel(data vendor) model.Vendor {
	v := model.Vendor{
		Name:                      data.Name,
		Email:                     data.Email,
		Phone:                     data.Phone,
		CompanyName:               data.CompanyName,
		Description:               data.Description,
		Image:                     data.Image,
		Street1:                   data.Street1,
		Street2:                   data.Street2,
		City:                      data.City,
		State:                     data.State,
		Zip:                       data.Zip,
		Website:                   data.Website,
		Disabled:                  true,
		SocialFacebook:            data.SocialFacebook,
		SocialInstagram:           data.SocialInstagram,
		SocialEtsy:                data.SocialEtsy,
		SocialWebsite:             data.SocialWebsite,
		SocialBohoSite:            data.SocialBohoSite,
		DefaultPaymentMethodLast4: data.DefaultPaymentMethodLast4,
		StripeID:                  data.StripeID,
		Rejected:                  data.Rejected,
		StripeConnectedAccountID:  data.StripeConnectedAccountID,
		ConnectionApplied:         data.ConnectionApplied,
	}
	v.ID = data.ID
	return v
}

type vendor struct {
	ID                        uint       `json:"id"`
	Name                      string     `json:"name"`
	Email                     string     `json:"email"`
	Phone                     string     `json:"phone"`
	CompanyName               string     `json:"companyName"`
	Description               string     `json:"description"`
	Image                     string     `json:"image"`
	Street1                   string     `json:"street1"`
	Street2                   string     `json:"street2"`
	City                      string     `json:"city"`
	State                     string     `json:"state"`
	Zip                       string     `json:"zip"`
	Website                   string     `json:"website"`
	Disabled                  bool       `json:"disabled"`
	SocialFacebook            string     `json:"socialFacebook"`
	SocialInstagram           string     `json:"socialInstagram"`
	SocialEtsy                string     `json:"socialEtsy"`
	SocialWebsite             string     `json:"socialWebsite"`
	SocialBohoSite            string     `json:"socialBohoSite"`
	DefaultPaymentMethodLast4 string     `json:"defaultPaymentMethodLast4"`
	Categories                []category `json:"categories"`
	StripeID                  string     `json:"stripeID"`
	MailingList               bool       `json:"mailingList"`
	Rejected                  bool       `json:"rejected"`
	StripeConnectedAccountID  string     `json:"stripeConnectedAccountID"`
	ConnectionApplied         bool       `json:"connectionApplied"`
}
