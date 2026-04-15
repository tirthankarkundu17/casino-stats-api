package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
	log    *zap.SugaredLogger
}

// ConnectMongo establishes a connection to MongoDB
func ConnectMongo(uri, dbName string, log *zap.SugaredLogger) (*MongoInstance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	log.Info("Connected to MongoDB")

	return &MongoInstance{
		Client: client,
		Db:     client.Database(dbName),
		log:    log,
	}, nil
}

// Close closes the MongoDB connection
func (m *MongoInstance) Close() {
	if err := m.Client.Disconnect(context.Background()); err != nil {
		m.log.Errorf("Failed to disconnect MongoDB: %v", err)
	}
}
