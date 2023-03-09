package storage

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Tx struct {
	Sign          string
	TxHash        string
	WalletAddress string
	Amount        uint64
	Lt            uint64
	Acked         bool
	CreatedAt     time.Time
}

type Donation struct {
	TxHash        string  `json:"txHash,omitempty" bson:"tx_hash,omitempty"`
	Sign          string  `json:"sign,omitempty" bson:"sign,omitempty"`
	WalletAddress string  `json:"wallet_address,omitempty" bson:"wallet_address,omitempty"`
	Amount        float64 `json:"amount,omitempty" bson:"amount,omitempty"`
	From          string  `json:"nickname,omitempty" bson:"nickname,omitempty"`
	StreamerId    string  `json:"streamerId,omitempty" bson:"streamer_id,omitempty"`
	Message       string  `json:"text,omitempty" bson:"message,omitempty"`
	Lt            string  `json:"lt,omitempty" bson:"lt,omitempty"`
	Verified      bool    `json:"verified,omitempty" bson:"verified,omitempty"`
	Acked         bool    `json:"acked,omitempty" bson:"acked,omitempty"`

	// ToDo: Add createdAt/modifiedAt
}

func (m *MongoStorage) GetDonationBySign(ctx context.Context, sign string) (*Donation, error) {
	dbName := os.Getenv("DB_NAME")
	collectionName := os.Getenv("DB_DONATIONS_COLLECTION_NAME")
	filter := bson.D{{Key: "sign", Value: sign}}
	// opts := options.FindOneOptions().SetSort(bson.D{{}})

	result := m.client.Database(dbName).Collection(collectionName).FindOne(ctx, filter)
	if result.Err() == mongo.ErrNoDocuments {
		// Return since no such document in mongo
		return nil, nil
	}

	var donation Donation
	err := result.Decode(&donation)
	if err != nil {
		return nil, err
	}

	return &donation, nil
}

func (m *MongoStorage) CreateDonation(ctx context.Context, donation Donation) (*mongo.InsertOneResult, error) {
	dbName := os.Getenv("DB_NAME")
	collectionName := os.Getenv("DB_DONATIONS_COLLECTION_NAME")

	result, err := m.client.Database(dbName).Collection(collectionName).InsertOne(ctx, donation)
	if err != nil {
		return nil, err
	}

	return result, nil
}
