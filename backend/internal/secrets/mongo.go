package secrets

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Reads secrets from a MongoDB collection (and can upload them)
// Expected document shape (one document per environment):
//
//	{ _id: "dev", dotenv: "KEY=value\nOTHER=value\n", updated_at: <date> }
type MongoProvider struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type secretDocument struct {
	ID     string `bson:"_id"`
	Dotenv string `bson:"dotenv"`
}

// Creates a provider
// mongo.Connect does not contact the server
// Never log the uri
func NewMongoProvider(uri, database, collection string) (*MongoProvider, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, errors.New("could not create secrets store client (check SECRETS_MONGO_URI)")
	}
	return &MongoProvider{
		client:     client,
		collection: client.Database(database).Collection(collection),
	}, nil
}

// Returns the stored .env text for an environment
func (p *MongoProvider) Fetch(ctx context.Context, environment string) (string, error) {
	var doc secretDocument
	err := p.collection.FindOne(ctx, bson.M{"_id": environment}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", fmt.Errorf("no secrets stored for environment %q", environment)
	}
	if err != nil {
		return "", fmt.Errorf("fetch secrets: %w", err)
	}
	return doc.Dotenv, nil
}

// Disconnects from MongoDB
func (p *MongoProvider) Close(ctx context.Context) error {
	return p.client.Disconnect(ctx)
}

// Stores dotenv text for an environment, creating or replacing its document
// Not part of Provider interface because the running server only reads
func (p *MongoProvider) Push(ctx context.Context, environment, dotenv string) error {
	replacement := bson.M{
		"dotenv":     dotenv,
		"updated_at": time.Now().UTC(),
	}
	_, err := p.collection.ReplaceOne(
		ctx,
		bson.M{"_id": environment},
		replacement,
		options.Replace().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("store secrets: %w", err)
	}
	return nil
}
