package token

import (
	"io/ioutil"
	"os"

	"github.com/dgrijalva/jwt-go"
)

type User struct {
	Name string `json:"name"`
}

type CustomClaims struct {
	User User `json:"user"`
	jwt.StandardClaims
}

type Decoder interface {
	Decode(tokenString string) (*CustomClaims, error)
}

type Encoder interface {
	Encode(user User) (string, error)
}

func LoadCertFromFile(fileName string) ([]byte, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}
