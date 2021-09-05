package controller

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"hawker/model"
)

type Category struct {
	db *gorm.DB
}

func NewCategoryController(d *gorm.DB) *Category {
	return &Category{
		db: d,
	}
}

func (h *Category) Edit(c *gin.Context) {
	var data category
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	//data validation
	v := model.Category{
		Name:        data.Name,
		Description: data.Description,
	}
	v.ID = data.ID
	h.db.Save(&v)
	c.JSON(http.StatusOK, categoryTranslation(v))
}

func (h *Category) Get(c *gin.Context) {
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
	var ve model.Category
	h.db.First(&ve, id)
	if ve.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such application",
		})
		return
	}
	c.JSON(http.StatusOK, categoryTranslation(ve))
	return
}

func (h *Category) Delete(c *gin.Context) {
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
	var v model.Category
	v.ID = uint(id)
	h.db.Delete(&v)
	c.JSON(http.StatusOK, gin.H{
		"message": "category deleted",
	})
	return
}

func (h *Category) List(c *gin.Context) {
	var ve []model.Category
	h.db.Find(&ve)
	var out []category
	for _, v := range ve {
		out = append(out, categoryTranslation(v))
	}
	c.JSON(http.StatusOK, out)
}

func categoryTranslation(e model.Category) category {
	return category{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
	}
}

type category struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
