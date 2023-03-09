package storage

import (
	"context"
	"fmt"
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
	// connectionString := fmt.Sprintf(
	// 	"mongodb://%s:%s@localhost:27017",
	// 	"root",
	// 	"rootpassword")

	fmt.Println("START CONNECTION")
	// "mongodb://root:rootpassword@127.0.0.1:27017/?authSource=admin&readPreference=primary"
	client, err := mongo.NewClient(options.Client().ApplyURI(connection))
	fmt.Println("COOLL CONNECTION")

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return nil, err
	}

	// ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	// defer cancel()

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
