package token

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	test_key = []byte("testsecret")
)

type User struct {
	Name string
}

type CustomClaims struct {
	User User
	jwt.StandardClaims
}

type TokenService struct {
	expirationPeriod time.Duration
}

func NewTokenService(ep time.Duration) *TokenService {
	return &TokenService{
		expirationPeriod: ep,
	}
}

// Decode a token string into a custom claims object
func (ts *TokenService) Decode(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return test_key, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid token: unable to parse custom claims")
	}

	return claims, nil
}

// Encode a user object into a JWT string
func (ts *TokenService) Encode(user User) (string, error) {
	expireToken := time.Now().Add(ts.expirationPeriod).Unix()

	claims := CustomClaims{
		user,
		jwt.StandardClaims{
			ExpiresAt: expireToken,
			Issuer:    "filesharing.auth.service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(test_key)
}
