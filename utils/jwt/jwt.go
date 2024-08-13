package jwt

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("votre_clé_secrète")

type Claims struct {
	UserId string `json:"userid"`
	jwt.StandardClaims
}

// GénérerJWT génère un nouveau JWT token
func GenerateJWT(userId string) (string, error) {
	expirationTime := time.Now().Add(72 * time.Hour)
	claims := &Claims{
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// DécoderJWT décode un JWT token
func DecoderJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("JWT invalide")
	}

	return claims, nil
}
