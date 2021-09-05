package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/invoice"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"

	"hawker/model"
	"hawker/pkg/Applications"
)

type ApplicationController struct {
	db         *gorm.DB
	stripeKey  string
	appService *Applications.Applications
}

func NewApplicationController(d *gorm.DB, sk string, a *Applications.Applications) *ApplicationController {
	stripe.Key = sk
	return &ApplicationController{
		db:         d,
		stripeKey:  sk,
		appService: a,
	}
}

/* Create
admin ability to create an application to pending status, bypassing the submit process (and app fee)
*/
func (h *ApplicationController) Create(c *gin.Context) {
	var data application
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	//TODO: data validation so we can give reasonable errors
	app, err := h.appService.CreateApplication(data.EventID, data.VendorID, data.SpaceType, data.Split, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}
	c.JSON(http.StatusOK, applicationTranslation(app))
}

func (h *ApplicationController) Get(c *gin.Context) {
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
	var ve model.Application
	h.db.First(&ve, id)
	if ve.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such application",
		})
		return
	}
	c.JSON(http.StatusOK, applicationTranslation(ve))
	return
}

func (h *ApplicationController) List(c *gin.Context) {
	//handle filters
	var eventID, userID int
	var err error
	var u model.User
	var vert model.Vendor
	eid := c.Query("eventID")
	if eid != "" {
		eventID, err = strconv.Atoi(eid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid event ID",
			})
			return
		}
	}
	uid := c.Query("userID")
	if uid != "" {
		userID, err = strconv.Atoi(uid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid user ID",
			})
			return
		}
		h.db.First(&u, userID)
		if u.ID < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid user ID",
			})
			return
		}
		h.db.First(&vert, "email = ? ", u.Email)
		if vert.ID < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid vendor ID",
			})
			return
		}
	}

	var ve []model.Application
	q := h.db.Preload("Event").Preload("Vendor")
	if eventID > 0 {
		q.Where("event_id = ?", eventID)
	}
	if vert.ID > 0 {
		q.Where("vendor_id = ?", vert.ID)
	}
	q.Find(&ve)
	var out []application
	for _, v := range ve {
		out = append(out, applicationTranslation(v))
	}
	c.JSON(http.StatusOK, out)
}

//TODO: filters
func (h *ApplicationController) ListSelf(c *gin.Context) {
	email, ok := c.Get("email")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "user not included",
		})
		return
	}
	var vendor model.Vendor
	h.db.First(&vendor, "email = ?", email)

	var ve []model.Application
	h.db.Preload("Event").Find(&ve, "vendor_id = ?", vendor.ID)
	var out []application
	for _, v := range ve {
		out = append(out, applicationTranslation(v))
	}
	c.JSON(http.StatusOK, out)
}

/* Submit
end users can submit an application
*/
func (h *ApplicationController) Submit(c *gin.Context) {
	type te struct {
		SpaceType string `json:"spaceType"`
		EventID   uint   `json:"eventID"`
		Split     bool   `json:"split"`
	}
	var data te
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	email, ok := c.Get("email")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "user not included",
		})
		return
	}
	var ve model.Vendor
	h.db.First(&ve, "email = ?", email)
	var ap model.Application
	h.db.First(&ap, "vendor_id = ? AND event_id = ?", ve.ID, data.EventID)
	if ap.ID > 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "you have already applied to this event",
		})
		return
	}

	app, err := h.appService.CreateApplication(data.EventID, ve.ID, data.SpaceType, data.Split, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, applicationTranslation(app))
}

/* SetStatus
handles changing application statuses
*/
func (h *ApplicationController) SetStatus(c *gin.Context) {
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
	var ve model.Application
	h.db.First(&ve, id)
	if ve.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such application",
		})
		return
	}
	appStatus := c.Param("status")
	if !h.appService.CanSetStatus(ve.Status, appStatus) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Cannot set status %s on app in status %s", appStatus, ve.Status),
		})
		return
	}
	app, err := h.appService.SetStatus(ve, appStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}
	c.JSON(http.StatusOK, applicationTranslation(app))
}

func (h *ApplicationController) GetInvoices(c *gin.Context) {
	idstr := c.Param("applicationID")
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
	var ap model.Application
	h.db.First(&ap, id)
	if ap.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "application not found",
		})
		return
	}
	var invoices []invoiceInfo

	if ap.InvoiceID != "" {
		in, err := invoice.Get(
			ap.InvoiceID,
			nil,
		)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "invoice not retreived",
			})
			return
		}

		invoices = append(invoices, invoiceTranslation(*in))
	}
	if ap.Invoice2ID != "" {
		in, err := invoice.Get(
			ap.Invoice2ID,
			nil,
		)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "invoice not retreived",
			})
			return
		}
		invoices = append(invoices, invoiceTranslation(*in))
	}

	c.JSON(http.StatusOK, invoices)
}

func applicationTranslation(a model.Application) application {
	app := application{
		ID:         a.ID,
		EventID:    a.EventID,
		VendorID:   a.VendorID,
		Status:     a.Status,
		Event:      eventTranslation(a.Event),
		Vendor:     vendorTranslation(a.Vendor),
		SpaceType:  a.SpaceType,
		Split:      a.Split,
		InvoiceID:  a.InvoiceID,
		Invoice2ID: a.Invoice2ID,
	}
	return app
}

//User view of an application
//should I make an admin and a non-admin one?
type application struct {
	ID         uint   `json:"id"`
	EventID    uint   `json:"eventID"`
	VendorID   uint   `json:"vendorID"`
	Status     string `json:"status"`
	SpaceType  string `json:"spaceType"`
	Split      bool   `json:"split"`
	Event      event  `json:"event"`
	Vendor     vendor `json:"vendor"`
	InvoiceID  string `json:"invoiceID"`
	Invoice2ID string `json:"invoice2ID"`
}

//invoice Details
//TODO: forgot to json these, so capitalization is broken front to backend here
type invoiceInfo struct {
	ID        string
	Status    string
	HostedURL string
	Total     int64
	Due       string
}

func invoiceTranslation(in stripe.Invoice) invoiceInfo {
	return invoiceInfo{
		ID:        in.ID,
		Status:    string(in.Status),
		HostedURL: in.HostedInvoiceURL,
		Total:     in.Total / 100,
		Due:       time.Unix(in.DueDate, 0).Format(time.RFC3339),
	}
}
