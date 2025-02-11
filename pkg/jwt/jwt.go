package jwt

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
)

const (
	tokenExpire     = 12 * time.Hour
	tokenSigningKey = "a1#slkj2327@KSmsda.#k32q"
)

type tokenClaims struct {
	jwt.StandardClaims
	UserID int `json:"id"`
}

func GenerateAccessToken(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenExpire).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userID,
	})

	signedToken, err := token.SignedString([]byte(tokenSigningKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ParseToken(accessToken string) (int, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing token method")
		}

		return []byte(tokenSigningKey), nil
	})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	return claims.UserID, nil
}
