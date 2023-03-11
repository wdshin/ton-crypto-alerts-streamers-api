package storage

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStorage struct {
	client *mongo.Client
}

func NewMongoClient(ctx context.Context) (*MongoStorage, error) {
	connection := os.Getenv("DB_CONNECTION")
	if connection == "" {
		connection = "mongodb://root:rootpassword@mongo:27017"
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(connection))

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return nil, err
	}

	if err = client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect using mongo client: %v", err)
		return nil, err
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping mongo: %v", err)
		return nil, err
	}

	return &MongoStorage{client: client}, nil
}
