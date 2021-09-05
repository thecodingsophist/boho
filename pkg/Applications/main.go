package Applications

import (
	"errors"
	"fmt"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/invoice"
	"gorm.io/gorm"
	"time"

	"hawker/model"
)

type Applications struct {
	db  *gorm.DB
	env string //running environment prod or ???
}

func NewApplications(d *gorm.DB, env string) *Applications {
	return &Applications{
		db:  d,
		env: env,
	}
}

/* CreateApplication
does the actual work of creating an application, invited or otherwise
*/
func (h *Applications) CreateApplication(eventID uint, vendorID uint, spaceType string, split bool, chargeAppFee bool) (model.Application, error) {
	app := model.Application{
		EventID:   eventID,
		VendorID:  vendorID,
		SpaceType: spaceType,
		Status:    model.ApplicationStatusPending,
		Split:     split,
	}
	var a model.Application
	h.db.First(&a, "event_id = ? AND vendor_id = ?", eventID, vendorID)
	if a.ID > 1 {
		return app, errors.New("application already exists")
	}

	err := h.db.Save(&app).Error
	if err != nil {
		fmt.Println(err)
		return app, err
	}

	if chargeAppFee {
		//charge the application fee... duh
		err = h.chargeApplicationFee(&app)
		if err != nil {
			return app, err
		}
	}
	return app, nil
}

/* SetStatus
Sets the status to the appropriate place.
*/
func (h *Applications) SetStatus(app model.Application, status string) (model.Application, error) {
	if !h.CanSetStatus(app.Status, status) {
		return app, errors.New("valid status not checked")
	}
	//2 statuses have extra work associated with them, both handle event space, and invoices
	if status == model.ApplicationStatusApproved {
		err := h.IssueInvoices(&app)
		if err != nil {
			return app, err
		}
		err = h.RemoveEventSpace(app.EventID, app.SpaceType)
		if err != nil {
			return app, err
		}
	} else if status == model.ApplicationStatusCancelled {
		err := h.AddEventSpace(app.EventID, app.SpaceType)
		if err != nil {
			return app, err
		}
		err = h.CancelInvoices(app)
		if err != nil {
			return app, err
		}
	}

	//only save the status change if everything worked
	app.Status = status
	h.db.Save(&app)
	return app, nil
}

/* CanSetStatus
Handles the rules for if the status can be set for an application
//First state: Pending
// Fee Failed -> Pending
// Pending -> WaitList, Approved, Rejected, MoreInfo
// WaitList -> Approved, Rejected
// MoreInfo -> Approved, Rejected
// Rejected -> nil
// Approved -> Cancelled, Paid, PartialPaid
// PartialPaid -> Cancelled, Paid
// Paid -> Cancelled, NoShow
// Cancelled -> Pending
*/
func (h *Applications) CanSetStatus(currentStatus string, newStatus string) bool {
	switch currentStatus {
	case model.ApplicationStatusPending:
		switch newStatus {
		case model.ApplicationStatusWaitList, model.ApplicationStatusApproved, model.ApplicationStatusRejected, model.ApplicationStatusMoreInfo:
			return true
		}
	case model.ApplicationStatusWaitList:
		switch newStatus {
		case model.ApplicationStatusApproved, model.ApplicationStatusRejected:
			return true
		}
	case model.ApplicationStatusMoreInfo:
		switch newStatus {
		case model.ApplicationStatusApproved, model.ApplicationStatusRejected:
			return true
		}
	case model.ApplicationStatusApproved:
		switch newStatus {
		case model.ApplicationStatusCancelled, model.ApplicationStatusPartialPaid, model.ApplicationStatusPaid:
			return true
		}
	case model.ApplicationStatusPartialPaid:
		switch newStatus {
		case model.ApplicationStatusCancelled, model.ApplicationStatusPaid:
			return true
		}
	case model.ApplicationStatusPaid:
		switch newStatus {
		case model.ApplicationStatusCancelled, model.ApplicationStatusNoShow:
			return true
		}
	case model.ApplicationStatusCancelled:
		switch newStatus {
		case model.ApplicationStatusPending:
			return true
		}
	case model.ApplicationStatusFeeFailed:
		switch newStatus {
		case model.ApplicationStatusPending:
			return true
		}
	}
	return false
}

/* AddEventSpace
adds back event space based on booth type (used when cancelling
*/
func (h *Applications) AddEventSpace(eventID uint, spaceType string) error {
	var eventInfo model.Event
	h.db.First(&eventInfo, eventID)
	switch spaceType {
	case model.EventSpaceSingle, model.EventSpaceFood:
		eventInfo.VendorSpaces += 1
	case model.EventSpaceDouble:
		eventInfo.VendorSpaces += 2
	case model.EventSpaceTruck:
		eventInfo.FoodTruckSpaces += 1
	}
	err := h.db.Save(&eventInfo).Error
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

/* RemoveEventSpace
removes event space when an application is approved
*/
func (h *Applications) RemoveEventSpace(eventID uint, spaceType string) error {
	var eventInfo model.Event
	h.db.First(&eventInfo, eventID)
	switch spaceType {
	case model.EventSpaceSingle, model.EventSpaceFood:
		eventInfo.VendorSpaces -= 1
	case model.EventSpaceDouble:
		eventInfo.VendorSpaces -= 2
	case model.EventSpaceTruck:
		eventInfo.FoodTruckSpaces -= 1
	}
	err := h.db.Save(&eventInfo).Error
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

/* CancelInvoices
cancels any unpaid invoices in relation to the application
*/
func (h *Applications) CancelInvoices(app model.Application) error {
	if app.Status == model.ApplicationStatusApproved {
		_, err := invoice.VoidInvoice(app.InvoiceID, nil)
		if err != nil {
			fmt.Println("could not cancel invoice: ", err)
			return err
		}
		if app.Split {
			_, err = invoice.VoidInvoice(app.Invoice2ID, nil)
			if err != nil {
				fmt.Println("could not cancel invoice: ", err)
				return err
			}
		}

	} else if app.Status == model.ApplicationStatusPartialPaid && app.Split{
		_, err := invoice.VoidInvoice(app.Invoice2ID, nil)
		if err != nil {
			fmt.Println("could not cancel invoice: ", err)
			return err
		}
	}
	return nil
}

/* IssueInvoices
does the hard work of determining what invoices to issue, and doing so
*/
func (h *Applications) IssueInvoices(app *model.Application) error {
	var vendor model.Vendor
	err := h.db.First(&vendor, app.VendorID).Error
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	var event model.Event
	err = h.db.First(&event, app.EventID).Error
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	price := getStripePrice(app.SpaceType, event)
	if price < 100 {
		fmt.Printf("price for event %d is under a dollar\n", event.ID)
		return errors.New("invalid price")
	}

	params := &stripe.InvoiceParams{
		Customer:         stripe.String(vendor.StripeID),
		CollectionMethod: stripe.String("send_invoice"),
		DueDate:          stripe.Int64(getDueDate(event.StartDate, true).Unix()),
	}

	itemParam := &stripe.InvoiceItemParams{
		Customer:    stripe.String(vendor.StripeID),
		Description: stripe.String(getItemDescription(event, app.SpaceType, app.Split, true)),
		PriceData: &stripe.InvoiceItemPriceDataParams{
			Currency:   stripe.String("USD"),
			Product:    stripe.String(h.getStripeProductID(app.SpaceType, false)),
			UnitAmount: stripe.Int64(price),
		},
	}

	//split invoices only pay half up front
	if app.Split {
		itemParam.PriceData.UnitAmount = stripe.Int64(price / 2)
	}
	invoiceID, err := createInvoice(params, itemParam)
	if err != nil {
		return err
	}
	app.InvoiceID = invoiceID
	h.db.Save(&app)

	if app.Split {
		params.DueDate = stripe.Int64(getDueDate(event.StartDate, false).Unix())
		itemParam.Description = stripe.String(getItemDescription(event, app.SpaceType, app.Split, false))
		itemParam.PriceData.Product = stripe.String(h.getStripeProductID(app.SpaceType, false))
		invoiceID, err = createInvoice(params, itemParam)
		if err != nil {
			return err
		}
		app.Invoice2ID = invoiceID
		h.db.Save(&app)
	}

	return nil
}

/* getDueDate
does the logic for invoice due dates
*/
func getDueDate(eventDate time.Time, first bool) time.Time {
	due := time.Now()
	if first {
		//first invoice is due 3 days from now, or day of event
		due = time.Now().Add(time.Hour * 24 * 3)
		if due.After(eventDate) {
			due = eventDate
		}
	} else {
		//the second is due 30 days before the event, or now, whichever is closer
		dueDate := eventDate.Add(time.Hour * 24 * -30)
		if !dueDate.Before(due) {
			due = dueDate
		}
	}

	//can never be due before now
	if due.Before(time.Now()) {
		due = time.Now().Add(time.Hour)
	}
	return due
}

/* getStripeProductID
gets the correct stripe product, based on environment also
*/
func (h *Applications) getStripeProductID(feeType string, first bool) string {
	if h.env == "prod" {
		if !first {
			return model.StripeRemainderSpaceFeeProd
		}
		switch feeType {
		case model.EventSpaceSingle:
			return model.StripeSingleSpaceFeeProd
		case model.EventSpaceDouble:
			return model.StripeDoubleSpaceFeeProd
		case model.EventSpaceFood:
			return model.StripeFoodSpaceFeeProd
		case model.EventSpaceTruck:
			return model.StripeTruckSpaceFeeProd
		case model.EventApplicationFee:
			return model.StripeApplicationFeeProd
		}
	}
	if !first {
		return model.StripeRemainderSpaceFee
	}
	switch feeType {
	case model.EventSpaceSingle:
		return model.StripeSingleSpaceFee
	case model.EventSpaceDouble:
		return model.StripeDoubleSpaceFee
	case model.EventSpaceFood:
		return model.StripeFoodSpaceFee
	case model.EventSpaceTruck:
		return model.StripeTruckSpaceFee
	case model.EventApplicationFee:
		return model.StripeApplicationFee
	}
	return ""
}

func (h *Applications) chargeApplicationFee(app *model.Application) error {
	var vendor model.Vendor
	err := h.db.First(&vendor, app.VendorID).Error
	if err != nil {
		fmt.Println(err)
		return err
	}
	var event model.Event
	err = h.db.First(&event, app.EventID).Error
	if err != nil {
		fmt.Println(err)
		return err
	}
	params := &stripe.InvoiceParams{
		Customer:         stripe.String(vendor.StripeID),
		CollectionMethod: stripe.String("charge_automatically"),
	}
	itemParam := &stripe.InvoiceItemParams{
		Customer:    stripe.String(vendor.StripeID),
		Description: stripe.String(getItemDescription(event, model.EventApplicationFee, false, false)),
		PriceData: &stripe.InvoiceItemPriceDataParams{
			Currency:   stripe.String("USD"),
			Product:    stripe.String(h.getStripeProductID(model.EventApplicationFee, false)),
			UnitAmount: stripe.Int64(getStripePrice(model.EventApplicationFee, event)),
		},
	}
	invoiceID, err := createInvoice(params, itemParam)
	if err != nil {
		return err
	}
	app.InvoiceFee = invoiceID
	h.db.Save(&app)

	_, err = invoice.Pay(invoiceID, nil)
	if err != nil {
		app.Status = model.ApplicationStatusFeeFailed
		h.db.Save(&app)
		fmt.Printf("could not pay invoice: %s\n", err.Error())
		return err
	}
	return nil
}
