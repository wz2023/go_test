package game_center

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"newstars/framework/model"
)

func VerifyCustomToken(tokenString string, secretKey []byte) (*model.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing error: %v", err)
	}
	if claims, ok := token.Claims.(*model.CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("token invalid")
}
