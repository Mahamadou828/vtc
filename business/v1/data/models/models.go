package models

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"vtc/business/v1/sys/database"
)

type Collection string

const (
	UserCollection  Collection = "user"
	RideCollection  Collection = "ride"
	OfferCollection Collection = "offer"
)

func Find[T any](ctx context.Context, client *mongo.Database, collectionName Collection, filter bson.D) ([]T, error) {
	res, err := database.Find[T](ctx, client, string(collectionName), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find %v: %v", collectionName, err)
	}

	return res, nil
}

func FindOne[T any](ctx context.Context, client *mongo.Database, collectionName Collection, filter bson.D) (*T, error) {
	var u *T

	if err := database.FindOne[T](ctx, client, string(collectionName), filter, u); err != nil {
		return nil, fmt.Errorf("failed to find one %v: %v", collectionName, err)
	}

	return u, nil
}

func InsertOne[T any](ctx context.Context, client *mongo.Database, collectionName Collection, u *T) error {
	if err := database.InsertOne[T](ctx, client, string(collectionName), u); err != nil {
		return fmt.Errorf("failed to insert one %v: %v", collectionName, err)
	}

	return nil
}

func InsertMany[T any](ctx context.Context, client *mongo.Database, collectionName Collection, dest []T) error {
	if err := database.InsertMany[T](ctx, client, string(collectionName), dest); err != nil {
		return fmt.Errorf("failed to inser many %vs: %v", collectionName, err)
	}

	return nil
}

func UpdateOne[T any](ctx context.Context, client *mongo.Database, collectionName Collection, id string, u *T) error {
	if err := database.UpdateOne[T](ctx, client, string(collectionName), id, u); err != nil {
		return fmt.Errorf("failed to update %v: %v", collectionName, err)
	}

	return nil
}

func DeleteOne[T any](ctx context.Context, client *mongo.Database, collectionName Collection, id string) error {
	if err := database.DeleteOne(ctx, client, string(collectionName), id); err != nil {
		return fmt.Errorf("failed to delete %v: %v", collectionName, err)
	}

	return nil
}
