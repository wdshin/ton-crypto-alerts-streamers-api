package storage

import (
	"context"
	"errors"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Widget struct {
	StreamerId    string  `json:"streamerId,omitempty" bson:"streamer_id,omitempty"`
	Type          string  `json:"type,omitempty" bson:"type,omitempty"`
	AmountGoal    float64 `json:"amount_goal,omitempty" bson:"amount_goal,omitempty"`
	AmountCurrent float64 `json:"amount_current,omitempty" bson:"amount_current,omitempty"`
}

func (m *MongoStorage) GetWidgets(ctx context.Context, streamerId string) (*[]Widget, error) {
	dbName := os.Getenv("DB_NAME")
	collectionName := os.Getenv("DB_WIDGETS_COLLECTION_NAME")

	// Check if streamer exist
	filter := bson.D{{Key: "streamer_id", Value: streamerId}}
	_, err := getStreamer(ctx, m.client, filter)
	if err != nil {
		return nil, err
	}

	// Load all widgets
	filter = bson.D{
		{Key: "streamer_id", Value: streamerId},
		// {Key: "is_completed", Value: false}, // ToDo: We will probably need also is completed flag for filtering
	}
	opts := options.Find().SetLimit(100)
	iter, err := m.client.Database(dbName).Collection(collectionName).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	var results []Widget
	if err := iter.All(ctx, &results); err != nil {
		return nil, errors.New("Failed to retrieve data!")
	}

	return &results, nil
}

func (m *MongoStorage) CreateWidget(ctx context.Context, widget Widget) (*mongo.InsertOneResult, error) {
	dbName := os.Getenv("DB_NAME")
	collectionName := os.Getenv("DB_WIDGETS_COLLECTION_NAME")

	// Check if streamer exist
	filter := bson.D{{Key: "streamer_id", Value: widget.StreamerId}}
	streamer, err := getStreamer(ctx, m.client, filter)
	if err != nil {
		return nil, err
	} else if streamer == nil {
		return nil, errors.New("Streamer does not exist.")
	}

	// Create new widget info
	result, err := m.client.Database(dbName).Collection(collectionName).InsertOne(ctx, widget)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m *MongoStorage) AddToCurrentAmount(ctx context.Context, streamerId string, donatedAmount uint64) (*mongo.UpdateResult, error) {
	dbName := os.Getenv("DB_NAME")
	collectionName := os.Getenv("DB_WIDGETS_COLLECTION_NAME")

	// Check if streamer exist
	filter := bson.D{{Key: "streamer_id", Value: streamerId}}
	_, err := getStreamer(ctx, m.client, filter)
	if err != nil {
		return nil, err
	}

	opts := options.Update().SetUpsert(true)
	filter = bson.D{{Key: "streamer_id", Value: streamerId}}
	update := bson.D{{Key: "$inc", Value: bson.D{
		{Key: "amount_current", Value: donatedAmount}}}}

	result, err := m.client.Database(dbName).Collection(collectionName).UpdateOne(ctx, filter, update, opts)

	if err != nil {
		return nil, err
	}

	return result, nil
}
