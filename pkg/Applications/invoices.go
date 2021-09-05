package Applications

import (
	"errors"
	"fmt"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/invoice"
	"github.com/stripe/stripe-go/v71/invoiceitem"

	"hawker/model"
)

/* createInvoice
actually does the work to create a stripe invoice
*/
func createInvoice(params *stripe.InvoiceParams, itemParam *stripe.InvoiceItemParams) (string, error) {
	var invoiceID string
	_, err := invoiceitem.New(itemParam)
	if err != nil {
		fmt.Printf("could not add invoice item: %s\n", err.Error())
		return invoiceID, errors.New("could not add invoice item")
	}
	in, err := invoice.New(params)
	if err != nil {
		fmt.Printf("could not generate invoice: %s\n", err.Error())
		return invoiceID, errors.New("could not generate invoice")
	}

	_, err = invoice.FinalizeInvoice(in.ID, nil)
	if err != nil {
		fmt.Printf("could not finalize invoice: %s\n", err.Error())
		return invoiceID, errors.New("could not finalize invoice")
	}
	invoiceID = in.ID
	return invoiceID, nil
}

/* getStripePrice
event prices are stored in dollars, but stripe expects prices in cents
*/
func getStripePrice(feeType string, event model.Event) int64 {
	switch feeType {
	case model.EventSpaceSingle:
		return int64(event.PriceSingleSpace) * 100
	case model.EventSpaceDouble:
		return int64(event.PriceDoubleSpace) * 100
	case model.EventSpaceFood:
		return int64(event.PriceFoodSpace) * 100
	case model.EventSpaceTruck:
		return int64(event.PriceTruckSpace) * 100
	case model.EventApplicationFee:
		return int64(500) //todo: change this from hardcoded to event based
	}

	return 0
}

/* getItemDescription
does string building to make a descriptive invoice item
*/
func getItemDescription(event model.Event, feeType string, split, first bool) string {
	feeName := "Application fee"
	switch feeType {
	case model.EventSpaceSingle:
		feeName = "Single Space"
	case model.EventSpaceDouble:
		feeName = "Double Space"
	case model.EventSpaceFood:
		feeName = "Food Space"
	case model.EventSpaceTruck:
		feeName = "Truck Space"
	}
	if feeType != model.EventApplicationFee {
		predicate := "Full "
		if split {
			predicate = "Remainder "
		}
		if first {
			predicate = "First "
		}
		feeName = predicate + feeName
	}
	return fmt.Sprintf("%s for %s on %s", feeName, event.Name, event.StartDate.Format("January 2, 2006"))
}

/* UpdateFullInvoices
checks non-split approved invoices for payments
*/
func (h *Applications) UpdateFullInvoices() (updated uint) {

	var apps []model.Application
	h.db.Find(&apps, "status = ? and split = false", "approved")
	for _, app := range apps {
		//each app in the list, get the stripe invoice, and mark as paid if its paid
		in, err := invoice.Get(app.InvoiceID, nil)
		if err != nil {
			fmt.Printf("could not get invoice %s, err %s\n", app.InvoiceID, err.Error())
			continue
		}
		if in.Status == stripe.InvoiceStatusPaid {
			_, err = h.SetStatus(app, model.ApplicationStatusPaid)
			if err != nil {
				continue
			}
			updated++
		}
	}
	return
}

/* UpdateFirstInvoices
updates the first invoice for split payments
*/
func (h *Applications) UpdateFirstInvoices() (updated uint) {
	var apps []model.Application
	h.db.Find(&apps, "status = ? and split = true", "approved")
	for _, app := range apps {
		//each app in the list, get the stripe invoice, and mark as paid if its paid
		in, err := invoice.Get(app.InvoiceID, nil)
		if err != nil {
			fmt.Printf("could not get invoice %s, err %s\n", app.InvoiceID, err.Error())
			continue
		}
		if in.Status == stripe.InvoiceStatusPaid {
			_, err = h.SetStatus(app, model.ApplicationStatusPartialPaid)
			if err != nil {
				continue
			}
			updated++
		}
	}
	return
}

/* UpdateFinalInvoices
updates split invoices that are already partially paid
*/
func (h *Applications) UpdateFinalInvoices() (updated uint) {
	var apps []model.Application
	h.db.Find(&apps, "status = ? and split = true", "partial")
	for _, app := range apps {
		//each app in the list, get the stripe invoice, and mark as paid if its paid
		in, err := invoice.Get(app.Invoice2ID, nil)
		if err != nil {
			fmt.Printf("could not get invoice %s, err %s\n", app.Invoice2ID, err.Error())
			continue
		}
		if in.Status == stripe.InvoiceStatusPaid {
			_, err = h.SetStatus(app, model.ApplicationStatusPaid)
			if err != nil {
				continue
			}
			updated++
		}
	}
	return
}
