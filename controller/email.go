package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/mattbaird/gochimp"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"

	"hawker/model"
)

type Email struct {
	db       *gorm.DB
	client   *gochimp.MandrillAPI
	from     string
	fromName string
}

func NewEmailController(d *gorm.DB) *Email {
	//TODO: move to env variable
	client, err := gochimp.NewMandrill("SPs3Pg3HRMAK0W3BSGspJA")
	if err != nil {
		log.Print("failed to do mandrill")
	}
	return &Email{
		db:       d,
		client:   client,
		from:     "info@thebohomarket.co",
		fromName: "The Boho Market",
	}
}

func (h *Email) GetTemplate(c *gin.Context) {
	template := c.Param("template")
	var ret model.EmailTemplate
	h.db.First(&ret, "email_type = ?", template)
	c.JSON(http.StatusOK, emailTemplate{
		EmailType: ret.EmailType,
		Subject:   ret.Subject,
		Body:      ret.Body,
	})
}

func (h *Email) SendEmail(c *gin.Context) {
	var e email
	err := c.ShouldBind(&e)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	var toSet []string
	toSet = append(toSet, e.To)
	h.send(toSet, e.Subject, e.Body)
	c.JSON(http.StatusOK, gin.H{})
}

func (h *Email) Broadcast(c *gin.Context) {
	var e email
	err := c.ShouldBind(&e)
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
	var applications []model.Application
	h.db.Preload("Vendor").Find(&applications, "status = ?", model.ApplicationStatusPaid)
	var toSet []string
	for _, a := range applications {
		toSet = append(toSet, a.Vendor.Email)
	}
	h.send(toSet, e.Subject, e.Body)
	c.JSON(http.StatusOK, gin.H{})
}

func (h *Email) List(c *gin.Context) {
	var emails []model.EmailTemplate
	var ret []emailTemplate
	h.db.Find(&emails)
	for _, e := range emails {
		ret = append(ret, emailTemplate{
			EmailType: e.EmailType,
			Subject:   e.Subject,
			Body:      e.Body,
		})
	}
	c.JSON(http.StatusOK, ret)
}

func (h *Email) Update(c *gin.Context) {
	var data emailTemplate
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	var old model.EmailTemplate
	h.db.First(&old, "email_type = ?", data.EmailType)
	if old.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid template",
		})
		return
	}
	old.Subject = data.Subject
	old.Body = data.Body
	h.db.Save(&old)
	c.JSON(http.StatusOK, emailTemplate{
		EmailType: old.EmailType,
		Subject:   old.Subject,
		Body:      old.Body,
	})
}

type emailTemplate struct {
	EmailType string `json:"emailType"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

type email struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func (h *Email) send(tos []string, subject, body string) {
	var recipients []gochimp.Recipient
	for _, t := range tos {
		recipients = append(recipients, gochimp.Recipient{
			Email: t,
		})
	}
	message := gochimp.Message{
		Subject:   subject,
		Text:      body,
		FromEmail: h.from,
		FromName:  h.fromName,
		To:        recipients,
	}

	_, _ = h.client.MessageSend(message, false)
}
