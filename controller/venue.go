package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"hawker/model"
)

type Venue struct {
	db *gorm.DB
}

func NewVenueController(d *gorm.DB) *Venue {
	return &Venue{db: d}
}
func (h *Venue) Edit(c *gin.Context) {
	var data venue
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	//data validation
	v := venueToModel(data)
	h.db.Save(&v)
	data.ID = v.ID

	var vcat []model.VenueCategory
	h.db.Delete(&vcat, "venue_id = ?", v.ID)

	for _, cat := range data.Categories {
		ca := model.VenueCategory{
			VenueID:    v.ID,
			CategoryID: cat.ID,
		}
		vcat = append(vcat, ca)
	}
	h.db.Save(&vcat)

	out, err := h.getVenue(data.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no such venue found",
		})
		return
	}

	c.JSON(http.StatusOK, venueTranslation(out))
}

func (h *Venue) Get(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid id",
		})
		return
	}
	if id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid id",
		})
		return
	}
	out, err := h.getVenue(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no such venue found",
		})
		return
	}

	c.JSON(http.StatusOK, venueTranslation(out))
}

func (h *Venue) Delete(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid id",
		})
		return
	}
	if id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid id",
		})
		return
	}
	var venue model.Venue
	venue.ID = uint(id)
	h.db.Delete(&venue)
	c.JSON(http.StatusOK, gin.H{
		"message": "venue deleted",
	})
	return
}

func (h *Venue) List(c *gin.Context) {
	var venues []model.Venue
	h.db.Find(&venues)
	var out []venue
	for _, v := range venues {
		out = append(out, venueTranslation(v))
	}
	c.JSON(http.StatusOK, out)
}

//does a fresh pull of things for a vendor
func (h *Venue) getVenue(id uint) (model.Venue, error) {
	var ve model.Venue
	h.db.Preload("Categories").Preload("Categories.Category").First(&ve, id)
	if ve.ID < 1 {
		return ve, errors.New("no such venue")
	}
	return ve, nil
}

func venueTranslation(data model.Venue) venue {
	vr := venue{
		ID:               data.ID,
		Name:             data.Name,
		Street1:          data.Street1,
		Street2:          data.Street2,
		City:             data.City,
		State:            data.State,
		Zip:              data.Zip,
		Email:            data.Email,
		Phone:            data.Phone,
		Hours:            data.Hours,
		FAQ:              data.FAQ,
		AllowFood:        data.AllowFood,
		AllowFoodTruck:   data.AllowFoodTruck,
		VendorSpaces:     data.VendorSpaces,
		FoodTruckSpaces:  data.FoodTruckSpaces,
		PriceSingleSpace: data.PriceSingleSpace,
		PriceDoubleSpace: data.PriceDoubleSpace,
		PriceFoodSpace:   data.PriceFoodSpace,
		PriceTruckSpace:  data.PriceTruckSpace,
		EventGuide:       data.EventGuide,
		BoothLayout:      data.BoothLayout,
	}
	for _, cat := range data.Categories {
		ca := category{
			ID:          cat.CategoryID,
			Name:        cat.Category.Name,
			Description: cat.Category.Description,
		}
		vr.Categories = append(vr.Categories, ca)
	}
	return vr
}

func venueToModel(data venue) model.Venue {
	ev := model.Venue{
		Name:             data.Name,
		Street1:          data.Street1,
		Street2:          data.Street2,
		City:             data.City,
		State:            data.State,
		Zip:              data.Zip,
		Email:            data.Email,
		Phone:            data.Phone,
		Hours:            data.Hours,
		FAQ:              data.FAQ,
		AllowFood:        data.AllowFood,
		AllowFoodTruck:   data.AllowFoodTruck,
		VendorSpaces:     data.VendorSpaces,
		FoodTruckSpaces:  data.FoodTruckSpaces,
		PriceSingleSpace: data.PriceSingleSpace,
		PriceDoubleSpace: data.PriceDoubleSpace,
		PriceFoodSpace:   data.PriceFoodSpace,
		PriceTruckSpace:  data.PriceTruckSpace,
		EventGuide:       data.EventGuide,
		BoothLayout:      data.BoothLayout,
	}
	ev.ID = data.ID
	return ev
}

type venue struct {
	ID               uint       `json:"id"`
	Name             string     `json:"name"`
	Street1          string     `json:"street1"`
	Street2          string     `json:"street2"`
	City             string     `json:"city"`
	State            string     `json:"state"`
	Zip              string     `json:"zip"`
	Email            string     `json:"email"`
	Phone            string     `json:"phone"`
	Hours            string     `json:"hours"`
	FAQ              string     `json:"faq"`
	AllowFood        bool       `json:"allowFood"`
	AllowFoodTruck   bool       `json:"allowFoodTruck"`
	VendorSpaces     uint       `json:"vendorSpaces"`
	FoodTruckSpaces  uint       `json:"foodTruckSpaces"`
	PriceSingleSpace uint       `json:"priceSingleSpace"`
	PriceDoubleSpace uint       `json:"priceDoubleSpace"`
	PriceFoodSpace   uint       `json:"priceFoodSpace"`
	PriceTruckSpace  uint       `json:"priceTruckSpace"`
	EventGuide       string     `json:"eventGuide"`
	BoothLayout      string     `json:"boothLayout"`
	Categories       []category `json:"categories"`
}
