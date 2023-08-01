package ride

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"vtc/business/v1/sys/database"
)

const (
	collectionName = "offer"
)

func Find(ctx context.Context, client *mongo.Database, filter bson.D) ([]Ride, error) {
	res, err := database.Find[Ride](ctx, client, collectionName, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find offer: %v", err)
	}

	return res, nil
}

func FindOne(ctx context.Context, client *mongo.Database, filter bson.D) (Ride, error) {
	var u Ride

	if err := database.FindOne[Ride](ctx, client, collectionName, filter, &u); err != nil {
		return Ride{}, fmt.Errorf("failed to find one offer: %v", err)
	}

	return u, nil
}

func InsertOne(ctx context.Context, client *mongo.Database, u Ride) error {
	if err := database.InsertOne[Ride](ctx, client, collectionName, u); err != nil {
		return fmt.Errorf("failed to insert one offer: %v", err)
	}

	return nil
}

func InsertMany(ctx context.Context, client *mongo.Database, offers []Ride) error {
	if err := database.InsertMany[Ride](ctx, client, collectionName, offers); err != nil {
		return fmt.Errorf("failed to inser many offers: %v", err)
	}

	return nil
}

func UpdateOne(ctx context.Context, client *mongo.Database, id string, u Ride) error {
	if err := database.UpdateOne[Ride](ctx, client, collectionName, id, u); err != nil {
		return fmt.Errorf("failed to update offer: %v", err)
	}

	return nil
}

func DeleteOne(ctx context.Context, client *mongo.Database, id string) error {
	if err := database.DeleteOne(ctx, client, collectionName, id); err != nil {
		return fmt.Errorf("failed to delete offer: %v", err)
	}

	return nil
}
