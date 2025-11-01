package mongo

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aluo/gomono/zeonology/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DB_NAME = "zeonology"
)

// Returns new mongo client
func NewMongoClient(cfg *config.Config) (*mongo.Client, error) {

	if cfg.Mongo.Uri == "" {
		return nil, errors.New("the mongo uri is not set")
	}
	log.Printf("Mongo Uri: %s", cfg.Mongo.Uri)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Mongo.Uri))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func Close(client *mongo.Client) {
	if err := client.Disconnect(context.Background()); err != nil {
		panic(err)
	}
}

func CreateTimeSeriesCollection(client *mongo.Client, collecName string) (*mongo.Collection, error) {
	db := client.Database(DB_NAME)

	// Creates a time series collection that stores "price" values over time
	tso := options.TimeSeries().SetTimeField("timestamp")
	opts := options.CreateCollection().SetTimeSeriesOptions(tso)

	err := db.CreateCollection(context.TODO(), collecName, opts)
	if err != nil {
		// Check if the error is not due to the collection already existing
		if mongoErr, ok := err.(mongo.CommandError); ok && mongoErr.Code != 48 {
			// Handle other errors
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}
		fmt.Println("Collection already exists, ignoring the error.")
	}

	collection := db.Collection(collecName)
	return collection, nil
}

func CreateCollection(client *mongo.Client, collecName string) (*mongo.Collection, error) {
	db := client.Database(DB_NAME)
	collec := db.Collection(collecName)
	return collec, nil
}
