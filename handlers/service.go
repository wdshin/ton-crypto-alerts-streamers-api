package handlers

import (
	"net/http"

	"github.com/vladtenlive/ton-donate/storage"
	"github.com/vladtenlive/ton-donate/utils"
)

type Service struct {
	client       *http.Client
	storage      storage.Storage
	mongoStorage *storage.MongoStorage
	auth         *utils.Auth
}

func NewService(client *http.Client, storage storage.Storage, mongoStorage *storage.MongoStorage, auth *utils.Auth) *Service {
	return &Service{
		client:       client,
		storage:      nil,
		mongoStorage: mongoStorage,
		auth:         auth,
	}
}
