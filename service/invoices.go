package service

import (
	"fmt"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/invoice"
	"gorm.io/gorm"

	"hawker/model"
)

type Invoice struct {
	db        *gorm.DB
	stripeKey string
}

func NewInvoice(d *gorm.DB, sk string) *Invoice {
	stripe.Key = sk
	return &Invoice{
		db:        d,
		stripeKey: sk,
	}
}

func (h *Invoice) CheckInvoices() {
	var apps []model.Application
	h.db.Find(&apps, "status = ? and split = false", "approved")
	h.saveFullInvoices(apps)
	h.db.Find(&apps, "status = ? and split = true", "partial")
	h.save2ndInvoices(apps)
	h.db.Find(&apps, "status = ? and split = true", "approved")
	h.savePartialInvoices(apps)
}

func (h *Invoice) saveFullInvoices(apps []model.Application) {
	for _, app := range apps {
		//each app in the list, get the stripe invoice, and mark as paid if its paid
		in, err := invoice.Get(app.InvoiceID, nil)
		if err != nil {
			fmt.Printf("could not get invoice %s, err %s\n", app.InvoiceID, err.Error())
			continue
		}
		if in.Status == stripe.InvoiceStatusPaid {
			app.Status = model.ApplicationStatusPaid
			h.db.Save(&app)
		}
	}
}

func (h *Invoice) savePartialInvoices(apps []model.Application) {
	for _, app := range apps {
		//each app in the list, get the stripe invoice, and mark as paid if its paid
		in, err := invoice.Get(app.InvoiceID, nil)
		if err != nil {
			fmt.Printf("could not get invoice %s, err %s\n", app.InvoiceID, err.Error())
			continue
		}
		if in.Status == stripe.InvoiceStatusPaid {
			app.Status = model.ApplicationStatusPartialPaid
			h.db.Save(&app)
		}
	}
}

func (h *Invoice) save2ndInvoices(apps []model.Application) {
	for _, app := range apps {
		//each app in the list, get the stripe invoice, and mark as paid if its paid
		in, err := invoice.Get(app.Invoice2ID, nil)
		if err != nil {
			fmt.Printf("could not get invoice %s, err %s\n", app.Invoice2ID, err.Error())
			continue
		}
		if in.Status == stripe.InvoiceStatusPaid {
			app.Status = model.ApplicationStatusPaid
			h.db.Save(&app)
		}
	}
}
