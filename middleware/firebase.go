package middleware

import (
	"errors"
	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"

	"hawker/model"
)

type FirebaseAuth struct {
	authClient *auth.Client
	db         *gorm.DB
}

func NewFirebaseMiddleware(a *auth.Client, d *gorm.DB) *FirebaseAuth {
	return &FirebaseAuth{
		authClient: a,
		db:         d,
	}
}

func (h *FirebaseAuth) ValidateJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		const BEARER = "bearer"
		at := c.GetHeader("Authorization")
		if len(at) <= len(BEARER)+1 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		token := at[len(BEARER):]
		decode, err := h.authClient.VerifyIDToken(c, token)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		var u model.User
		err = h.db.First(&u, "uid = ?", decode.Claims["user_id"]).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//this is a user I dont know, save them in the user table at least so I can do somethign about it
			u.UID = decode.Claims["user_id"].(string)
			u.Email = decode.Claims["email"].(string)
			//had to disable a user here too... probably need to think about breaking out the user creation thing at some point
			u.Disabled = true
			h.db.Save(&u)
		}
		c.Set("uid", decode.Claims["user_id"])
		c.Set("email", decode.Claims["email"])
		c.Set("isAdmin", decode.Claims["isAdmin"])
		if u.IsAdmin {
			c.Set("isAdmin", true)
		}
	}
}
