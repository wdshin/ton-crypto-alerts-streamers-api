package utils

import (
	"context"
	"errors"
	"fmt"
	"log"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/jwk"
)

// Auth ...
type Auth struct {
	jwk               *JWK
	jwkURL            string
	cognitoRegion     string
	cognitoUserPoolID string
	publicKeySet      jwk.Set
}

// Config ...
type Config struct {
	CognitoRegion     string
	CognitoUserPoolID string
}

// JWK ...
type JWK struct {
	Keys []KeyType `json:"keys"`
}

type KeyType struct {
	Alg string `json:"alg"`
	E   string `json:"e"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	Use string `json:"use"`
}

// NewAuth ...
func NewAuth(ctx context.Context, config *Config) *Auth {
	a := &Auth{
		cognitoRegion:     config.CognitoRegion,
		cognitoUserPoolID: config.CognitoUserPoolID,
		jwkURL:            fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", config.CognitoRegion, config.CognitoUserPoolID),
	}

	publicKeySet, err := a.LoadAndCachePublicKeyPair(ctx)
	if err != nil {
		log.Fatal(err)
	}

	a.publicKeySet = publicKeySet

	return a
}

func (a *Auth) LoadAndCachePublicKeyPair(ctx context.Context) (jwk.Set, error) {
	// "kid" must be present in the public keys set
	publicKeySet, err := jwk.Fetch(ctx, a.jwkURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to load public key set")
	}

	return publicKeySet, nil
}

// Parse jwt
func (a *Auth) ParseJWT(tokenString string) (*jwt.Token, *jwt.MapClaims, error) {
	var userClaims *jwt.MapClaims = &jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, userClaims, func(token *jwt.Token) (interface{}, error) {

		// Verify if the token was signed with correct signing method
		// AWS Cognito is using RSA256 in my case
		_, ok := token.Method.(*jwt.SigningMethodRSA)

		if !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// Get "kid" value from token header
		// "kid" is shorthand for Key ID
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid header not found")
		}

		key, found := a.publicKeySet.LookupKeyID(kid)
		if !found {
			return nil, fmt.Errorf("unable to find key %q", kid)
		}
		// if len(keys) == 0 {
		// 	return nil, fmt.Errorf("key %v not found", kid)
		// }

		// In our case, we are returning only one key = keys[0]
		// Return token key as []byte{string} type
		var tokenKey interface{}
		if err := key.Raw(&tokenKey); err != nil {
			return nil, errors.New("failed to create token key")
		}

		return tokenKey, nil
	})

	return token, userClaims, err
}
