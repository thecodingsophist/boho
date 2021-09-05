package model

import (
	"gorm.io/gorm"
)

type EmailTemplate struct {
	gorm.Model
	EmailType string
	Subject   string
	Body      string
}

const (
	EmailTypeApprove   = "Approve"
	EmailTypeWaitlist  = "WaitList"
	EmailTypeMoreInfo  = "MoreInfo"
	EmailTypeReject    = "Reject"
	EmailTypeBroadcast = "Broadcast"
)
