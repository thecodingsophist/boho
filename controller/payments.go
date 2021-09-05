package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/customer"
	"github.com/stripe/stripe-go/v71/invoice"
	"github.com/stripe/stripe-go/v71/loginlink"
	"github.com/stripe/stripe-go/v71/paymentmethod"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"

	"hawker/model"
	"hawker/pkg/Applications"
)

type Payment struct {
	db         *gorm.DB
	stripeKey  string
	appService *Applications.Applications
}

func NewPaymentController(d *gorm.DB, sk string, a *Applications.Applications) *Payment {
	stripe.Key = sk
	return &Payment{
		db:         d,
		stripeKey:  sk,
		appService: a,
	}
}

func (h *Payment) SavePaymentMethod(c *gin.Context) {
	email, ok := c.Get("email")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid email",
		})
		return
	}
	var ve model.Vendor
	h.db.First(&ve, "email = ?", email)
	if ve.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("vendor %s does not exist", email),
		})
		return
	}
	type token struct {
		Token string
	}
	var s token
	err := c.ShouldBind(&s)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "bad input",
		})
		return
	}
	var cust *stripe.Customer
	src := &stripe.SourceParams{Token: stripe.String(s.Token)}
	cp := &stripe.CustomerParams{Source: src}

	if len(ve.StripeID) < 1 {
		cp.Email = stripe.String(ve.Email)
		cp.Name = stripe.String(ve.Name)
		cp.Phone = stripe.String(ve.Phone)
		cust, err = customer.New(cp)
		if err != nil {
			fmt.Print("could not create: ", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "customer creation failed",
			})
			return
		}
		ve.StripeID = cust.ID
		h.db.Save(&ve)
	} else {
		cust, err = customer.Update(ve.StripeID, cp)
		if err != nil {
			fmt.Println("could not update: ", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "could not set payment method: " + err.Error(),
			})
			return
		}
	}

	if cust == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not handle stripe customer",
		})
		return
	}

	var cardSet []Cards
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(ve.StripeID),
		Type:     stripe.String("card"),
	}
	i := paymentmethod.List(params)
	for i.Next() {
		pm := i.PaymentMethod()
		cr := Cards{
			Last4:    pm.Card.Last4,
			ExpMonth: pm.Card.ExpMonth,
			ExpYear:  pm.Card.ExpYear,
		}

		if cust.DefaultSource != nil && cust.DefaultSource.ID == pm.ID {
			cr.IsDefault = true
			ve.DefaultPaymentMethodLast4 = cr.Last4
			h.db.Save(&ve)
		}
		cardSet = append(cardSet, cr)
	}
	c.JSON(http.StatusOK, cardSet)
}

func (h *Payment) GetPaymentMethods(c *gin.Context) {
	email, ok := c.Get("email")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid email",
		})
		return
	}
	var ve model.Vendor
	h.db.First(&ve, "email = ?", email)
	if ve.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such vendor",
		})
		return
	}
	var cardSet []Cards
	if len(ve.StripeID) < 1 {
		c.JSON(http.StatusOK, cardSet)
		return
	}
	if ve.StripeID == "" {
		c.JSON(http.StatusOK, cardSet)
	}
	cust, err := customer.Get(ve.StripeID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(ve.StripeID),
		Type:     stripe.String("card"),
	}
	i := paymentmethod.List(params)
	for i.Next() {
		pm := i.PaymentMethod()
		cr := Cards{
			Last4:    pm.Card.Last4,
			ExpMonth: pm.Card.ExpMonth,
			ExpYear:  pm.Card.ExpYear,
		}
		if cust.DefaultSource.ID == pm.ID {
			cr.IsDefault = true
		}
		cardSet = append(cardSet, cr)
	}
	c.JSON(http.StatusOK, cardSet)
}

func (h *Payment) UpdateInvoices(c *gin.Context) {
	u1 := h.appService.UpdateFullInvoices()
	u2 := h.appService.UpdateFirstInvoices()
	u3 := h.appService.UpdateFinalInvoices()
	c.JSON(http.StatusOK, gin.H{
		"fullInvoices":  u1,
		"firstInvoices": u2,
		"finalInvoices": u3,
	})
}

func (h *Payment) ListInvoices(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid id number",
		})
		return
	}
	if id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid id",
		})
		return
	}
	var ap model.User
	h.db.First(&ap, id)
	if ap.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "user not found",
		})
		return
	}
	var ve model.Vendor
	h.db.First(&ve, "email = ?", ap.Email)
	if ap.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "vendor not found",
		})
		return
	}
	if len(ve.StripeID) < 1 {
		c.JSON(http.StatusOK, nil)
		return
	}
	type invoiceInfo struct {
		ID        string
		Status    string
		HostedURL string
		Total     int64
		Due       string
	}
	var invoices []invoiceInfo
	params := &stripe.InvoiceListParams{
		Customer: stripe.String(ve.StripeID),
	}
	raw := invoice.List(params)
	for raw.Next() {
		in := raw.Invoice()
		invoices = append(invoices, invoiceInfo{
			ID:        in.ID,
			Status:    string(in.Status),
			HostedURL: in.HostedInvoiceURL,
			Total:     in.Total / 100,
			Due:       time.Unix(in.DueDate, 0).Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, invoices)
}

func (h *Payment) GetLoginLink(c *gin.Context) {
	a, ok := c.Get("email")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "no email in client",
		})
		return
	}
	email := a.(string)
	var ve model.Vendor
	h.db.First(&ve, "email = ?", email)
	if ve.StripeConnectedAccountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such account",
		})
		return
	}
	ll, err := loginlink.New(&stripe.LoginLinkParams{Account: stripe.String(ve.StripeConnectedAccountID)})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"url": ll.URL,
	})
}

type Cards struct {
	Last4     string `json:"last4"`
	ExpMonth  uint64 `json:"expMonth"`
	ExpYear   uint64 `json:"expYear"`
	IsDefault bool   `json:"isDefault"`
}
