package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"

	jwt2 "hawker/pkg/jwt"
)

type AuthHandler struct {
	jwtService *jwt2.Handler
}

func NewAuthHandler(j *jwt2.Handler) *AuthHandler {
	return &AuthHandler{jwtService: j}
}

func (h *AuthHandler) ValidateJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		const BEARER = "bearer"
		auth := c.GetHeader("Authorization")
		if len(auth) < len(BEARER) {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		token := auth[len(BEARER):]
		decode, _ := h.jwtService.Validate(token)
		if !decode.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		//TODO: add claims to context so I know who it is and if they are an admin? or just decode on the fly?
		claims := decode.Claims.(jwt.MapClaims)
		c.Set("email", claims["email"])
		c.Set("isAdmin", claims["isAdmin"])
	}
}
