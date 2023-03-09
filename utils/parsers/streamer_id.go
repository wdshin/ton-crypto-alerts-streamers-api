package utils

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func GetStreamerId(r *http.Request) string {
	paramName := "streamerId"

	if headerValue := r.Header.Get("Authorization"); headerValue != "" {
		// ToDo: Parse header value here and get cognitoId
		cognitoId := ""
		return cognitoId
	} else if cognitoId := chi.URLParam(r, paramName); cognitoId != "" {
		return cognitoId
	} else if cognitoId := r.URL.Query().Get(paramName); cognitoId != "" {
		return cognitoId
	} else {
		return ""
	}
}

func GetCognitoId(r *http.Request) string {
	paramName := "cognitoId"

	if headerValue := r.Header.Get("Authorization"); headerValue != "" {
		// ToDo: Parse header value here and get cognitoId
		cognitoId := ""
		return cognitoId
	} else if cognitoId := chi.URLParam(r, paramName); cognitoId != "" {
		return cognitoId
	} else if cognitoId := r.URL.Query().Get(paramName); cognitoId != "" {
		return cognitoId
	} else {
		return ""
	}
}
