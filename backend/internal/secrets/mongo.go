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

var (
	// ErrMongoClientCreation is returned when the MongoDB client cannot be created.
	ErrMongoClientCreation = errors.New("could not create secrets store client (check SECRETS_MONGO_URI)")
	// ErrSecretsNotFound is returned when no secrets are stored for an environment.
	ErrSecretsNotFound = errors.New("no secrets stored for environment")
)

// MongoProvider reads secrets from a MongoDB collection (and can upload them).
// Expected document shape (one document per environment):
//
//	{ _id: "dev", dotenv: "KEY=value\nOTHER=value\n", updated_at: <date> }
type MongoProvider struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type secretDocument struct {
	Dotenv string `bson:"dotenv"`
}

// NewMongoProvider creates a provider.
// mongo.Connect does not contact the server.
// Never log the uri.
func NewMongoProvider(uri, database, collection string) (*MongoProvider, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMongoClientCreation, err)
	}

	return &MongoProvider{
		client:     client,
		collection: client.Database(database).Collection(collection),
	}, nil
}

// Fetch returns the stored .env text for an environment.
func (p *MongoProvider) Fetch(ctx context.Context, environment string) (string, error) {
	var doc secretDocument

	err := p.collection.FindOne(ctx, bson.M{"_id": environment}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", fmt.Errorf("%w %q", ErrSecretsNotFound, environment)
	}

	if err != nil {
		return "", fmt.Errorf("fetch secrets: %w", err)
	}

	return doc.Dotenv, nil
}

// Close disconnects from MongoDB.
func (p *MongoProvider) Close(ctx context.Context) error {
	return p.client.Disconnect(ctx)
}

// Push stores dotenv text for an environment, creating or replacing its document.
// Not part of Provider interface because the running server only reads.
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
