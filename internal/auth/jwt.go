package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func ParseRSAPublicKeyFromPEM(pubPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pubPEM)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("invalid public key PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not RSA public key")
	}
	return rsaPub, nil
}

// VerifyAndParseJWT tekshiradi va token ichidagi claimsni qaytaradi
func VerifyAndParseJWT(tokenString string, pubKey *rsa.PublicKey) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Algoritm tekshirish
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return pubKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
