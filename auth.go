package main

import (
	"errors"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// AuthService provides authentication service
type AuthService interface {
	Auth(string, string) (string, error)
}

type authService struct {
	key     []byte
	clients map[string]string
}

type customClaims struct {
	ClientID string `json:"clientId"`
	jwt.StandardClaims
}

const expiration = 120

func generateToken(signingKey []byte, clientID string) (string, error) {
	claims := customClaims{
		clientID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * expiration).Unix(),
			IssuedAt:  jwt.TimeFunc().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey)
}

func (as authService) Auth(clientID string, clientSecret string) (string, error) {
	if as.clients[clientID] == clientSecret {
		signed, err := generateToken(as.key, clientID)
		if err != nil {
			return "", errors.New(err.Error())
		}
		return signed, nil
	}
	return "", ErrAuth
}

// ErrAuth is returned when credentials are incorrect
var ErrAuth = errors.New("Incorrect credentials")
