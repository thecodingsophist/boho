package controller

import (
	"context"
	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"hawker/model"
	"hawker/service"
)

type User struct {
	db          *gorm.DB
	authClient  *auth.Client
	userService *service.User
}

func NewUserController(d *gorm.DB, a *auth.Client, u *service.User) *User {
	return &User{
		db:          d,
		authClient:  a,
		userService: u,
	}
}

func (h *User) Get(c *gin.Context) {
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
	var ve model.User
	h.db.First(&ve, id)
	if ve.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such vendor",
		})
		return
	}
	ven, _ := h.userService.GetVendorByEmail(ve.Email)
	c.JSON(http.StatusOK, userTranslation(ve, ven))
	return
}

func (h *User) Update(c *gin.Context) {
	var v user
	err := c.ShouldBind(&v)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
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
	var ve model.User
	h.db.First(&ve, id)
	if ve.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such vendor",
		})
		return
	}
	//data validation
	ve = model.User{
		Email:     v.Email,
		UID:       v.UID,
		IsAdmin:   v.IsAdmin,
		Onboarded: v.Onboarded,
		Disabled:  v.Disabled,
	}
	ve.ID = uint(id)
	h.db.Save(&ve)
	ven, _ := h.userService.GetVendorByEmail(ve.Email)
	c.JSON(http.StatusOK, userTranslation(ve, ven))
}

func (h *User) Enable(c *gin.Context) {
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
	var ve model.User
	h.db.First(&ve, id)
	if ve.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such vendor",
		})
		return
	}
	ve.Disabled = false
	h.db.Save(&ve)
	ven, _ := h.userService.GetVendorByEmail(ve.Email)
	c.JSON(http.StatusOK, userTranslation(ve, ven))
}

func (h *User) Disable(c *gin.Context) {
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
	var ve model.User
	h.db.First(&ve, id)
	if ve.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such user",
		})
		return
	}
	ve.Disabled = true
	h.db.Save(&ve)
	ven, _ := h.userService.GetVendorByEmail(ve.Email)
	c.JSON(http.StatusOK, userTranslation(ve, ven))
}

func (h *User) List(c *gin.Context) {
	var ve []model.User
	var ver []model.Vendor
	var out []user
	vMap := make(map[string]model.Vendor)
	h.db.Find(&ve)
	ver, _ = h.userService.ListVendors()
	for _, ven := range ver {
		vMap[ven.Email] = ven
	}

	for _, v := range ve {
		out = append(out, userTranslation(v, vMap[v.Email]))
	}
	c.JSON(http.StatusOK, out)
}

func (h *User) AddAdmin(c *gin.Context) {
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
	var ve model.User
	h.db.First(&ve, id)
	if ve.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such user",
		})
		return
	}
	ve.IsAdmin = true
	claims := map[string]interface{}{"isAdmin": true}
	err = h.authClient.SetCustomUserClaims(context.Background(), ve.UID, claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	}
	h.db.Save(&ve)
	ven, _ := h.userService.GetVendorByEmail(ve.Email)
	c.JSON(http.StatusOK, userTranslation(ve, ven))
}

func (h *User) RemoveAdmin(c *gin.Context) {
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
	var ve model.User
	h.db.First(&ve, id)
	if ve.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such user",
		})
		return
	}
	ve.IsAdmin = false
	claims := map[string]interface{}{"isAdmin": false}
	err = h.authClient.SetCustomUserClaims(context.Background(), ve.UID, claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	}
	h.db.Save(&ve)
	ven, _ := h.userService.GetVendorByEmail(ve.Email)
	c.JSON(http.StatusOK, userTranslation(ve, ven))
}

//both how a user gets created, as well as how we get initial user details
func (h *User) GetOrCreate(c *gin.Context) {
	type info struct {
		Email string
		Name  string
	}
	var data info
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	uid := c.Param("uid")
	if len(uid) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid uid",
		})
	}
	if data.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid email",
		})
		return

	}
	var u model.User
	h.db.First(&u, "uid = ?", uid)
	if u.ID < 1 {
		u.Email = data.Email
		u.Disabled = true
		h.db.Save(&u)
	}
	ven, _ := h.userService.GetVendorByEmail(u.Email)
	c.JSON(http.StatusOK, userTranslation(u, ven))
}

func (h *User) Reject(c *gin.Context) {
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
	//Security check
	admin, ok := c.Get("isAdmin")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "admin key failure",
		})
		return
	}
	if !admin.(bool) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "only admins may modify vendor profiles",
		})
		return
	}
	var u model.User
	h.db.First(&u, id)
	if u.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such user",
		})
		return
	}
	var ve model.Vendor
	h.db.First(&ve, "email = ?", u.Email)
	if ve.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such vendor",
		})
		return
	}
	ve.Rejected = true
	h.db.Save(&ve)
	c.JSON(http.StatusOK, nil)
}

func (h *User) ApplyConnection(c *gin.Context) {
	email, ok := c.Get("email")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "user not included",
		})
		return
	}
	var ve model.Vendor
	h.db.First(&ve, "email = ?", email)

	ve.ConnectionApplied = true
	h.db.Save(&ve)
	c.JSON(http.StatusOK, nil)
}

func (h *User) UpdateStripeInfo(c *gin.Context) {
	type stripeInfo struct {
		StripeConnectedAccountID  string `json:"stripeConnectedAccountID"`
		StripeID                  string `json:"stripeID"`
		DefaultPaymentMethodLast4 string `json:"defaultPaymentMethodLast4"`
	}
	var v stripeInfo
	err := c.ShouldBind(&v)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
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
	//Security check
	admin, ok := c.Get("isAdmin")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "admin key failure",
		})
		return
	}
	if !admin.(bool) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "only admins may modify vendor profiles",
		})
		return
	}
	var u model.User
	h.db.First(&u, id)
	if u.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such user",
		})
		return
	}
	var ve model.Vendor
	h.db.First(&ve, "email = ?", u.Email)
	if ve.ID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no such vendor",
		})
		return
	}
	ve.StripeConnectedAccountID = v.StripeConnectedAccountID
	ve.StripeID = v.StripeID
	ve.DefaultPaymentMethodLast4 = v.DefaultPaymentMethodLast4
	h.db.Save(&ve)
	c.JSON(http.StatusOK, nil)
}

func userTranslation(v model.User, ve model.Vendor) user {
	return user{
		ID:        v.ID,
		Email:     v.Email,
		UID:       v.UID,
		IsAdmin:   v.IsAdmin,
		Onboarded: v.Onboarded,
		Disabled:  v.Disabled,
		Vendor:    vendorTranslation(ve),
	}
}

type user struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	UID       string `json:"uid"` //google firebase auth identifier
	IsAdmin   bool   `json:"isAdmin"`
	Onboarded bool   `json:"onboarded"`
	Disabled  bool   `json:"disabled"`
	Vendor    vendor `json:"vendor"`
}
