package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"sso/internal/domain/models"
	"time"
)

// NewToken генерация нового токета
func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodES256)
	claims := token.Claims.(jwt.MapClaims)

	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["app_id"] = app.ID

	//Подписываем свой токен
	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}
	return tokenString, err
}
