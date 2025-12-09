package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type JwtCustomClaims struct {
	CreateTime int64  `json:"createTime"`
	UserID     uint   `json:"userID"`
	Email      string `json:"email"`
	jwt.StandardClaims
}

func main() {
	now := time.Now()
	expiredAt := now.Add(time.Hour * 24).Unix()
	claims := &JwtCustomClaims{
		now.UnixNano() / int64(time.Millisecond),
		1,
		"test@example.com",
		jwt.StandardClaims{
			ExpiresAt: expiredAt,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		panic(err)
	}
	fmt.Println(accessToken)
}
