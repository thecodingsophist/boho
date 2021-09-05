package jwt

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type Handler struct {
	secret string
	issuer string
}

type tokenClaims struct {
	Email   string `json:"email"`
	IsAdmin bool   `json:"isAdmin"`
	jwt.StandardClaims
}

func NewHandler(secret, issuer string) *Handler {
	return &Handler{
		secret: secret,
		issuer: issuer,
	}
}

func (h *Handler) Validate(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			return nil, fmt.Errorf("invalid token", token.Header["alg"])

		}
		return []byte(h.secret), nil
	})
}

func (h *Handler) Generate(email string, isAdmin bool) (string, error) {
	claims := &tokenClaims{
		Email:   email,
		IsAdmin: isAdmin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "hawker",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.secret))

}
