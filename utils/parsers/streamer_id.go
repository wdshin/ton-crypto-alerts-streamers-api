package utils

import (
	"net/http"

	"github.com/vladtenlive/ton-donate/utils"
)

func GetStreamerId(r *http.Request, auth *utils.Auth) string {
	if headerValue := r.Header.Get("Authorization"); headerValue != "" {
		_, claims, err := auth.ParseJWT(headerValue)
		if err != nil {
			return ""
		}

		streamerId, err := claims.GetSubject()
		if err != nil {
			return ""
		}

		return streamerId
	} else {
		return ""
	}
}

func GetCognitoId(r *http.Request, auth *utils.Auth) string {
	if headerValue := r.Header.Get("Authorization"); headerValue != "" {
		_, claims, err := auth.ParseJWT(headerValue)
		if err != nil {
			return ""
		}

		cognitoId, err := claims.GetSubject() // Is same as sub
		if err != nil {
			return ""
		}

		return cognitoId
	} else {
		return ""
	}
}
