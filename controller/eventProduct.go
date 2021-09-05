package controller

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"hawker/model"
)

type Product struct {
	db *gorm.DB
}

func NewEventProductController(d *gorm.DB) *Product {
	return &Product{
		db: d,
	}
}

func (h *Product) Create(c *gin.Context) {
	var data product
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	//data validation
	v := model.Product{
		Name:            data.Name,
		Description:     data.Description,
		BasePrice:       data.BasePrice,
		StripeProductID: data.StripeProductID,
	}
	h.db.Save(&v)
	data.ID = v.ID
	c.JSON(http.StatusOK, productTranslation(v))
}

func (h *Product) Get(c *gin.Context) {
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
	var ve model.Product
	h.db.First(&ve, id)
	if ve.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such application",
		})
		return
	}
	c.JSON(http.StatusOK, productTranslation(ve))
	return
}

func (h *Product) Update(c *gin.Context) {
	var data product
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
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
	ev := model.Product{
		Name:            data.Name,
		Description:     data.Description,
		BasePrice:       data.BasePrice,
		StripeProductID: data.StripeProductID,
	}
	ev.ID = uint(id)
	h.db.Save(&ev)
	c.JSON(http.StatusOK, productTranslation(ev))
}

func (h *Product) Delete(c *gin.Context) {
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
	var v model.Product
	v.ID = uint(id)
	h.db.Delete(&v)
	c.JSON(http.StatusOK, gin.H{
		"message": "product deleted",
	})
	return
}

func (h *Product) List(c *gin.Context) {
	var ve []model.Product
	h.db.Find(&ve)
	var out []product
	for _, v := range ve {
		out = append(out, productTranslation(v))
	}
	c.JSON(http.StatusOK, out)
}

func productTranslation(e model.Product) product {
	return product{
		ID:              e.ID,
		Name:            e.Name,
		Description:     e.Description,
		BasePrice:       e.BasePrice,
		StripeProductID: e.StripeProductID,
	}
}

type product struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	BasePrice       uint   `json:"basePrice"`
	StripeProductID string `json:"stripeProductID"`
}
