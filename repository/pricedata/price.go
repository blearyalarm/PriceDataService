package pricedata

import (
	"context"
	"log"
	"time"

	helper "github.com/erich/pricetracking/helper/mongo"
	"github.com/erich/pricetracking/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DB_NAME         = "zeonology"
	COLLECTION_NAME = "priceData"
)

// Define a struct that matches the aggregation output
type AggregationResult struct {
	ID struct {
		Interval time.Time `bson:"interval"`
	} `bson:"_id"`
	AggValue float64 `bson:"aggValue"`
}

type priceDataMongoRepo struct {
	mongoClient *mongo.Client

	priceDataCollection *mongo.Collection
}

type PriceDataMongoRepo interface {
	Create(ctx context.Context, pg []model.Entry) error
	Find(ctx context.Context, query model.Query) ([]model.Entry, error)
}

func NewPriceDataMongoRepo(client *mongo.Client) (PriceDataMongoRepo, error) {
	collection, err := helper.CreateTimeSeriesCollection(client, COLLECTION_NAME)
	if err != nil {
		return nil, err
	}

	return &priceDataMongoRepo{
		mongoClient:         client,
		priceDataCollection: collection,
	}, nil
}

// Create implements PhotographerMongoRepo.
func (p *priceDataMongoRepo) Create(ctx context.Context, pg []model.Entry) error {
	newData := make([]interface{}, len(pg))
	for i, v := range pg {
		newData[i] = v
	}

	_, err := p.priceDataCollection.InsertMany(ctx, newData)
	if err != nil {
		return err
	}
	return nil
}

// FindByEventType implements PhotographerMongoRepo.
func (p *priceDataMongoRepo) Find(ctx context.Context, query model.Query) ([]model.Entry, error) {

	// Aggregation pipeline
	pipeline := mongo.Pipeline{
		// Match documents within the specified time range
		{{"$match", bson.D{
			{"timestamp", bson.D{
				{"$gte", query.StartTime},
				{"$lt", query.EndTime},
			}},
		}}},
		// Group data into 1-minute intervals and compute the average value
		{{"$group", bson.D{
			{"_id", bson.D{
				{"interval", bson.M{"$dateTrunc": bson.M{
					"date":    "$timestamp",
					"unit":    query.WindowUnit,
					"binSize": query.WindowInterval,
				}}},
			}},
			{"aggValue", bson.M{"$" + string(query.Aggregation): "$price"}},
		}}},
	}

	// Execute the aggregation query
	cursor, err := p.priceDataCollection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatalf("Failed to aggregate data: %v", err)
	}
	defer cursor.Close(ctx)

	var results []AggregationResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert aggregation results to model.Entry
	entries := make([]model.Entry, len(results))
	for i, result := range results {
		entries[i] = model.Entry{
			Time:  result.ID.Interval,
			Value: result.AggValue,
		}
	}

	return entries, nil
}
