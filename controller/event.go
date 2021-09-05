package controller

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"

	"hawker/model"
)

type Event struct {
	db            *gorm.DB
	storageClient *storage.Client
	bucketName    string //bucket to store event images in
}

func NewEventController(d *gorm.DB, sc *storage.Client, bn string) *Event {
	return &Event{
		db:            d,
		storageClient: sc,
		bucketName:    bn,
	}
}

func (h *Event) Edit(c *gin.Context) {
	var e event
	err := c.ShouldBind(&e)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	ev := eventToModel(e)
	h.db.Save(&ev)

	if e.Image != "" {
		//TODO: handle the image and assoc
		objectPath := "eventImages/"
		objectName := strconv.Itoa(int(ev.ID))
		var objectExtenstion string
		var im image.Image
		mime1 := e.Image[strings.IndexByte(e.Image, ':')+1:]
		mime1 = mime1[:strings.IndexByte(mime1, ';')]
		b64data := e.Image[strings.IndexByte(e.Image, ',')+1:]
		unbased, err := base64.StdEncoding.DecodeString(b64data)
		if err != nil {
			fmt.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to process image",
			})
			return
		}
		r := bytes.NewReader(unbased)
		//need type to convert to image.image
		//then can encode to a writer stream, but need the type for the encode function and object name
		switch mime1 {
		case "image/gif":
			objectExtenstion = ".gif"
			im, err = gif.Decode(r)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "failed to process gif",
				})
				return
			}
		case "image/png":
			objectExtenstion = ".png"
			im, err = png.Decode(r)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "failed to process png",
				})
				return
			}
		case "image/jpeg":
			objectExtenstion = ".jpg"
			im, err = jpeg.Decode(r)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "failed to process jpeg",
				})
				return
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "unknown image type: " + mime1,
			})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
		defer cancel()
		wc := h.storageClient.Bucket(h.bucketName).Object(objectPath + objectName + objectExtenstion).NewWriter(ctx)

		switch mime1 {
		case "image/gif":
			err = gif.Encode(wc, im, nil)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "failed to save gif",
				})
				return
			}
		case "image/png":
			err = png.Encode(wc, im)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "failed to save png",
				})
				return
			}

		case "image/jpeg":
			err = jpeg.Encode(wc, im, nil)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "failed to save jpeg",
				})
				return
			}
		}

		//todo: make thumb public
		//close client stream
		err = wc.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to finalize file",
			})
			return
		}

		//make image public
		acl := h.storageClient.Bucket(h.bucketName).Object(objectPath + objectName + objectExtenstion).ACL()
		if err := acl.Set(context.Background(), storage.AllUsers, storage.RoleReader); err != nil {
			fmt.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to publish object",
			})
			return
		}

		ev.ImageURL = objectPath + objectName + objectExtenstion
		//add image url to event
		h.db.Save(&ev)
	}

	var vcat []model.EventCategory
	h.db.Delete(&vcat, "event_id = ?", ev.ID)

	for _, cat := range e.Categories {
		ca := model.EventCategory{
			EventID:    ev.ID,
			CategoryID: cat.ID,
		}
		vcat = append(vcat, ca)
	}
	h.db.Save(&vcat)
	out, err := h.getEvents(ev.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no such event found",
		})
		return
	}

	c.JSON(http.StatusOK, eventTranslation(out))
}

func (h *Event) Get(c *gin.Context) {
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
	out, err := h.getEvents(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no such event found",
		})
		return
	}

	//check secruity
	if out.EventStatus != model.EventStatusOpen && !c.GetBool("isAdmin") {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "you do not have permission to see this event",
		})
	}

	c.JSON(http.StatusOK, eventTranslation(out))
}

func (h *Event) getEvents(id uint) (model.Event, error) {
	var ev model.Event
	h.db.Preload("Categories").Preload("Categories.Category").First(&ev, id)
	if ev.ID < 1 {
		return ev, errors.New("no such event")
	}
	return ev, nil
}

func (h *Event) Cancel(c *gin.Context) {
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
	var ev model.Event
	h.db.First(&ev, id)
	if ev.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid event",
		})
		return
	}
	ev.EventStatus = model.EventStatusCancelled
	h.db.Save(&ev)
	c.JSON(http.StatusOK, gin.H{
		"message": "event has been cancelled, note, currently this does not notify vendors or take any other action",
	})
}

//todo: admin level, no filters
func (h *Event) List(c *gin.Context) {
	var ve []model.Event
	fmt.Print(time.Now().String(), " start")
	h.db.Order("start_date asc").Find(&ve)
	fmt.Print(time.Now().String(), " query")
	var out []event
	for _, v := range ve {
		out = append(out, eventTranslation(v))
	}
	fmt.Print(time.Now().String(), " send")
	c.JSON(http.StatusOK, out)
	fmt.Print(time.Now().String(), " did")
	return
}

func (h *Event) ListOpen(c *gin.Context) {
	var ve []model.Event
	h.db.Preload("Venue").Order("start_date asc").Find(&ve, "start_date >= NOW() AND (event_status = ? OR event_status =?)", model.EventStatusOpen, model.EventStatusClosed)
	var out []event
	for _, v := range ve {
		out = append(out, eventTranslation(v))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Event) GetCSV(c *gin.Context) {
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
	var ev model.Event
	var sb strings.Builder
	h.db.Preload("Participants").Preload("Participants.Vendor").First(&ev, id)
	headers := "\"Company Name\", \"Vendor Email\", \"Application Status\" \n"
	sb.WriteString(headers)
	for _, v := range ev.Participants {
		switch v.Status {
		case model.ApplicationStatusRejected, model.ApplicationStatusPending, model.ApplicationStatusCancelled:
			continue
		}
		sb.WriteString(v.Vendor.CompanyName)
		sb.WriteString(",")
		sb.WriteString(v.Vendor.Email)
		sb.WriteString(",")
		sb.WriteString(v.Status)
		sb.WriteString("\n")
	}
	c.JSON(http.StatusOK, gin.H{"data": sb.String()})
}

func eventTranslation(e model.Event) event {
	vr := event{
		ID:               e.ID,
		Name:             e.Name,
		Description:      e.Description,
		Hours:            e.Hours,
		StartDate:        e.StartDate,
		EndDate:          e.EndDate,
		ImageURL:         e.ImageURL,
		FAQ:              e.FAQ,
		VenueID:          e.VenueID,
		EventStatus:      e.EventStatus,
		AllowFood:        e.AllowFood,
		AllowFoodTruck:   e.AllowFoodTruck,
		VendorSpaces:     e.VendorSpaces,
		FoodTruckSpaces:  e.FoodTruckSpaces,
		PriceSingleSpace: e.PriceSingleSpace,
		PriceDoubleSpace: e.PriceDoubleSpace,
		PriceFoodSpace:   e.PriceFoodSpace,
		PriceTruckSpace:  e.PriceTruckSpace,
		EventGuide:       e.EventGuide,
		BoothLayout:      e.BoothLayout,
		City:             e.Venue.City,
		State:            e.Venue.State,
	}
	for _, cat := range e.Categories {
		ca := category{
			ID:          cat.CategoryID,
			Name:        cat.Category.Name,
			Description: cat.Category.Description,
		}
		vr.Categories = append(vr.Categories, ca)
	}
	return vr

}

func eventToModel(e event) model.Event {
	ev := model.Event{
		Name:             e.Name,
		Description:      e.Description,
		Hours:            e.Hours,
		StartDate:        e.StartDate,
		EndDate:          e.EndDate,
		ImageURL:         e.ImageURL,
		FAQ:              e.FAQ,
		VenueID:          e.VenueID,
		EventStatus:      e.EventStatus,
		AllowFood:        e.AllowFood,
		AllowFoodTruck:   e.AllowFoodTruck,
		VendorSpaces:     e.VendorSpaces,
		FoodTruckSpaces:  e.FoodTruckSpaces,
		PriceSingleSpace: e.PriceSingleSpace,
		PriceDoubleSpace: e.PriceDoubleSpace,
		PriceFoodSpace:   e.PriceFoodSpace,
		PriceTruckSpace:  e.PriceTruckSpace,
		EventGuide:       e.EventGuide,
		BoothLayout:      e.BoothLayout,
	}
	if e.EventStatus == "" {
		e.EventStatus = model.EventStatusOpen
	}
	ev.ID = e.ID
	return ev
}

type event struct {
	ID               uint       `json:"id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Hours            string     `json:"hours"`
	StartDate        time.Time  `json:"startDate"`
	EndDate          time.Time  `json:"endDate"`
	Image            string     `json:"image"`
	ImageURL         string     `json:"imageURL"`
	FAQ              string     `json:"faq"`
	VenueID          uint       `json:"venueID"`
	EventStatus      string     `json:"eventStatus"`
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
	City             string     `json:"city"`
	State            string     `json:"state"`
}
