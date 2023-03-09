package handlers

import (
	"net/http"

	"github.com/vladtenlive/ton-donate/storage"
)

type Service struct {
	client       *http.Client
	storage      storage.Storage
	mongoStorage *storage.MongoStorage
}

func NewService(client *http.Client, storage storage.Storage, mongoStorage *storage.MongoStorage) *Service {
	return &Service{
		client:       client,
		storage:      nil,
		mongoStorage: mongoStorage,
	}
}
