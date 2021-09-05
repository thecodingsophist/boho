package model

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Application struct {
	gorm.Model
	EventID      uint
	VendorID     uint
	SpaceType    string
	Status       string
	ApprovalDate *time.Time
	Event        Event  `gorm:"ForeignKey:event_id"`
	Vendor       Vendor `gorm:"ForeignKey:vendor_id"`
	Split        bool
	InvoiceFee   string
	InvoiceID    string
	Invoice2ID   string
}

/*
We have products in stripe that are set for each type. They have a different price per event, but by linking them we can
use the products to identify information on the sale in the future.

Since invoices can be split, the Remainder product represents all secondary, no matter the space type.
*/
const (
	StripeApplicationFee    = "prod_IgH55RgS1qeDgi"
	StripeSingleSpaceFee    = "prod_IgH2s3tuaijnuI"
	StripeDoubleSpaceFee    = "prod_IgH3iS39q60TsV"
	StripeFoodSpaceFee      = "prod_IgH3fI93u2Ey0K"
	StripeTruckSpaceFee     = "prod_IgH4u16aUarSFj"
	StripeRemainderSpaceFee = "prod_IgH2WqBsMGBN54"

	StripeApplicationFeeProd    = "prod_JKd7RwWfkwELgU"
	StripeSingleSpaceFeeProd    = "prod_JIacvR32rNjudh"
	StripeDoubleSpaceFeeProd    = "prod_JIac5Ac3Jdse6T"
	StripeFoodSpaceFeeProd      = "prod_JIadVDtU3I5fdc"
	StripeTruckSpaceFeeProd     = "prod_JIad6jEUYjKptr"
	StripeRemainderSpaceFeeProd = "prod_JKd8uccXstHOkm"
)

const (
	EventSpaceSingle    = "10x10"
	EventSpaceDouble    = "20x10"
	EventSpaceFood      = "food"
	EventSpaceTruck     = "food truck"
	EventApplicationFee = "application fee"
)

const (
	ApplicationStatusPending     = "pending"
	ApplicationStatusWaitList    = "waitlist"
	ApplicationStatusApproved    = "approved"
	ApplicationStatusRejected    = "rejected"
	ApplicationStatusCancelled   = "cancelled"
	ApplicationStatusNoShow      = "noShow"
	ApplicationStatusMoreInfo    = "info"
	ApplicationStatusPaid        = "paid"
	ApplicationStatusPartialPaid = "partial"
	ApplicationStatusFeeFailed   = "feeFailed"
)

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

func (a *Application) SetStatus(newStatus string) error {
	if a.Status == ApplicationStatusPending {
		switch newStatus {
		case ApplicationStatusWaitList, ApplicationStatusApproved, ApplicationStatusRejected, ApplicationStatusMoreInfo:
			a.Status = newStatus
			return nil
		}
	} else if a.Status == ApplicationStatusWaitList {
		switch newStatus {
		case ApplicationStatusApproved, ApplicationStatusRejected:
			a.Status = newStatus
			return nil
		}
	} else if a.Status == ApplicationStatusMoreInfo {
		switch newStatus {
		case ApplicationStatusApproved, ApplicationStatusRejected:
			a.Status = newStatus
			return nil
		}
	} else if a.Status == ApplicationStatusApproved {
		switch newStatus {
		case ApplicationStatusCancelled, ApplicationStatusPartialPaid, ApplicationStatusPaid:
			a.Status = newStatus
			return nil
		}
	} else if a.Status == ApplicationStatusPartialPaid {
		switch newStatus {
		case ApplicationStatusCancelled, ApplicationStatusPaid:
			a.Status = newStatus
			return nil
		}
	} else if a.Status == ApplicationStatusPaid {
		switch newStatus {
		case ApplicationStatusCancelled, ApplicationStatusNoShow:
			a.Status = newStatus
			return nil
		}
	} else if a.Status == ApplicationStatusCancelled {
		switch newStatus {
		case ApplicationStatusPending:
			a.Status = newStatus
			return nil
		}
	}
	return errors.New(fmt.Sprintf("a %s application cannot be set to %s", a.Status, newStatus))
}
