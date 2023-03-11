package storage

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Streamer struct {
	StreamerId    string `json:"streamerId,omitempty" bson:"streamer_id,omitempty"`
	WalletAddress string `json:"wallet_address,omitempty" bson:"wallet_address,omitempty"`
	CognitoId     string `json:"cognito_id,omitempty" bson:"cognito_id,omitempty"`
}

// type StreamersRepository interface {
// 	GetStreamer(ctx context.Context, streamerId string) (Streamer, error)
// }

func (m *MongoStorage) GetStreamerByCognitoId(ctx context.Context, cognitoId string) (*Streamer, error) {
	filter := bson.D{{Key: "cognito_id", Value: cognitoId}}
	return getStreamer(ctx, m.client, filter)
}

func (m *MongoStorage) GetStreamerByStreamerId(ctx context.Context, streamerId string) (*Streamer, error) {
	filter := bson.D{{Key: "streamer_id", Value: streamerId}}
	return getStreamer(ctx, m.client, filter)
}

func (m *MongoStorage) GetStreamerByWalletAddress(ctx context.Context, walletAddress string) (*Streamer, error) {
	filter := bson.D{{Key: "wallet_address", Value: walletAddress}}
	return getStreamer(ctx, m.client, filter)
}

func getStreamer(ctx context.Context, client *mongo.Client, filter primitive.D) (*Streamer, error) {
	dbName := os.Getenv("DB_NAME")
	collectionName := os.Getenv("DB_STREAMERS_COLLECTION_NAME")

	var result Streamer
	docResult := client.Database(dbName).Collection(collectionName).FindOne(ctx, filter)
	if docResult.Err() == mongo.ErrNoDocuments {
		// Return since no such document in mongo
		return nil, nil
	}

	err := docResult.Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (m *MongoStorage) SaveStreamer(ctx context.Context, streamer Streamer) (*mongo.UpdateResult, error) {
	dbName := os.Getenv("DB_NAME")
	collectionName := os.Getenv("DB_STREAMERS_COLLECTION_NAME")

	opts := options.Update().SetUpsert(true)
	filter := bson.D{{Key: "streamer_id", Value: streamer.StreamerId}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "streamer_id", Value: streamer.StreamerId},
		{Key: "cognito_id", Value: streamer.CognitoId},
		{Key: "wallet_address", Value: streamer.WalletAddress}}}}

	result, err := m.client.Database(dbName).Collection(collectionName).UpdateOne(ctx, filter, update, opts)

	if err != nil {
		return nil, err
	}

	return result, nil
}
