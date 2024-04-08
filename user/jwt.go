package user

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"go_auth/config"
	"time"
)

func createToken(entity UserEntity) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"userId":   entity.userId,
			"username": entity.Username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString([]byte(config.Config().Auth.Secret))

	return tokenString, err
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Config().Auth.Secret), nil
	})

	if !token.Valid {
		return token, fmt.Errorf("invalid token")
	}

	return token, err
}
