package pricedata

import (
	"context"
	"fmt"
	"log"
	"time"

	helper "github.com/erich/pricetracking/helper/mongo"
	"github.com/erich/pricetracking/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	COLLECTION_NAME_LASTRETRIVAL = "lastretrival"
)

// LastUpdateRepo handles operations related to the last update entry
type lastUpdateRepo struct {
	collection *mongo.Collection
}

type LastUpdateMongoRepo interface {
	Update(ctx context.Context, last time.Time) error
	Get(ctx context.Context) (time.Time, error)
}

func NewLastRetrivalMongoRepo(client *mongo.Client) (LastUpdateMongoRepo, error) {
	collection, err := helper.CreateCollection(client, COLLECTION_NAME_LASTRETRIVAL)
	if err != nil {
		return nil, err
	}

	return &lastUpdateRepo{
		collection: collection,
	}, nil
}

// Update updates the lastUpdateTime entry in the collection
func (repo *lastUpdateRepo) Update(ctx context.Context, newTime time.Time) error {
	filter := bson.D{{}}
	update := bson.D{
		{"$set", bson.D{
			{"lastUpdateTime", newTime},
		}},
	}

	log.Printf("Filter: %+v, Update: %+v\n", filter, update)
	// Upsert: create the document if it does not exist
	_, err := repo.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to update last update time: %w", err)
	}
	return nil
}

// Get retrieves the lastUpdateTime entry from the collection
func (repo *lastUpdateRepo) Get(ctx context.Context) (time.Time, error) {
	var entry model.LastUpdateEntry
	err := repo.collection.FindOne(ctx, bson.D{}).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			_, err2 := repo.collection.InsertOne(ctx, model.LastUpdateEntry{LastUpdateTime: time.Time{}})
			if err2 != nil {
				return time.Time{}, err
			}
			return time.Time{}, nil // return Zero time
		}
		return time.Time{}, err
	}
	return entry.LastUpdateTime, nil
}
