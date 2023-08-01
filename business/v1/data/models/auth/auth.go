package auth

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"vtc/business/v1/sys/database"
)

const (
	CollectionName = "user"
)

func Find(ctx context.Context, client *mongo.Database, filter bson.D) ([]User, error) {
	res, err := database.Find[User](ctx, client, CollectionName, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	return res, nil
}

func FindOne(ctx context.Context, client *mongo.Database, filter bson.D) (User, error) {
	var u User

	if err := database.FindOne[User](ctx, client, CollectionName, filter, &u); err != nil {
		return User{}, fmt.Errorf("failed to find one user: %v", err)
	}

	return u, nil
}

func InsertOne(ctx context.Context, client *mongo.Database, u User) error {
	if err := database.InsertOne[User](ctx, client, CollectionName, u); err != nil {
		return fmt.Errorf("failed to insert one user: %v", err)
	}

	return nil
}

func InsertMany(ctx context.Context, client *mongo.Database, users []User) error {
	if err := database.InsertMany[User](ctx, client, CollectionName, users); err != nil {
		return fmt.Errorf("failed to inser many users: %v", err)
	}

	return nil
}

func UpdateOne(ctx context.Context, client *mongo.Database, id string, u User) error {
	if err := database.UpdateOne[User](ctx, client, CollectionName, id, u); err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	return nil
}

func DeleteOne(ctx context.Context, client *mongo.Database, id string) error {
	if err := database.DeleteOne(ctx, client, CollectionName, id); err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	return nil
}
