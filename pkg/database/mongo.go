package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scripts-management/internal/config"
)

func NewMongoClient(config *config.Config) (*mongo.Client, error) {
	clientOptions := options.Client().
		ApplyURI(config.MongoURI).
		SetMaxPoolSize(20).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(10 * time.Second).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second).
		SetSocketTimeout(10 * time.Second).
		SetHeartbeatInterval(10 * time.Second).
		SetLocalThreshold(10 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func ProvideMongoDB(client *mongo.Client) *mongo.Database {
	return client.Database("scripts_management")
}

func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("scripts_management").Collection(collectionName)
}

func CloseMongoClient(client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := client.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}
