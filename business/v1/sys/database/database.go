// Package database implement expose method to allow communication with a mongodb database
// using the standard mongo db library for go
// https://github.com/mongodb/mongo-go-driver
// https://www.mongodb.com/docs/drivers/go/current/usage-examples/deleteOne/
package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	caFilePath = "business/v1/sys/database/rds-combined-ca-bundle.pem"

	connectTimeout = 5
	queryTimeout   = 30
)

type Config struct {
	Username   string
	Password   string
	Host       string
	Port       string
	Database   string
	SSLEnabled bool
}

// NewClient open a new connection and return a client for the selected database.
// It accepts an aws session and a secret manager pool name to fetch the credentials to connect to the database.
func NewClient(cfg Config) (*mongo.Database, error) {
	q := make(url.Values)

	if cfg.SSLEnabled {
		q.Set("ssl", strconv.FormatBool(cfg.SSLEnabled))
		q.Set("ssl_ca_certs", "rds-combined-ca-bundle.pem")
		q.Set("replicaSet", "rs0")
		q.Set("readPreference", "secondaryPreferred")
		q.Set("retryWrites", "false")
	}

	//creating connection string
	connectionURI := url.URL{
		Scheme:   "mongodb",
		User:     url.UserPassword(cfg.Username, cfg.Password),
		Host:     fmt.Sprintf("%v:%v", cfg.Host, cfg.Port),
		RawQuery: q.Encode(),
	}

	clientOpt := options.Client().ApplyURI(connectionURI.String())

	//handling sll if enabled
	if cfg.SSLEnabled {
		tlsConfig, err := getCustomTLSConfig(caFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get TLS configuration: %v", err)
		}

		clientOpt = clientOpt.SetTLSConfig(tlsConfig)
	}

	client, err := mongo.NewClient(clientOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB cluster: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB cluster: %v", err)
	}

	return client.Database(cfg.Database), nil
}

// Find executes a search and return all matching document.
// The filter parameter must be a document containing query operators and can be used to select which documents are included in the result.
// It cannot be nil. An empty document (e.g. bson.D{}) should be used to include all documents.
func Find[T any](ctx context.Context, client *mongo.Database, collection string, filter bson.D) ([]T, error) {
	var res []T

	nCtx, cancel := context.WithTimeout(ctx, queryTimeout*time.Second)
	defer cancel()

	cur, err := client.Collection(collection).Find(nCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find collection with filter: %v, error: %v", filter, err)
	}

	defer cur.Close(nCtx)

	for cur.Next(nCtx) {
		var item T
		if err := cur.Decode(&item); err != nil {
			return nil, fmt.Errorf("failed to unmarshal item from result into destination: %v", err)
		}

		res = append(res, item)
	}

	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("failed to find in collection: %v", err)
	}

	return res, nil
}

// FindOne executes a search and return the first matching document.
// The filter parameter must be a document containing query operators and can be used to select which documents are included in the result.
// It cannot be nil. An empty document (e.g. bson.D{}) should be used to include all documents.
func FindOne[T any](ctx context.Context, client *mongo.Database, collection string, filter bson.D, dest *T) error {
	nCtx, cancel := context.WithTimeout(ctx, queryTimeout*time.Second)
	defer cancel()

	switch err := client.Collection(collection).FindOne(nCtx, filter).Decode(dest); err {
	case mongo.ErrNoDocuments:
		return fmt.Errorf("failed to find document with current filter: %v, error: %v", filter, err)
	case nil:
		break
	default:
		return fmt.Errorf("failed to find document error: %v", err)
	}

	return nil
}

// InsertOne execute an insert query to insert a single document.
// The data parameter must be the document to be inserted. It cannot be nil.
// If the document does not have an _id field when transformed into BSON, one will be added automatically to the marshalled document.
// The original document will not be modified.
func InsertOne[T any](ctx context.Context, client *mongo.Database, collection string, data *T) error {
	nCtx, cancel := context.WithTimeout(ctx, queryTimeout*time.Second)
	defer cancel()

	res, err := client.Collection(collection).InsertOne(nCtx, data)
	if err != nil {
		return fmt.Errorf("failed to insert document: %v", err)
	}

	print(res.InsertedID)
	return nil
}

// InsertMany executes an insert command to insert multiple documents.
// The data parameter must be a slice of documents to insert. The slice cannot be nil or empty. The elements must all be non-nil.
// For any document that does not have an _id field when transformed into BSON, one will be added automatically to the marshalled document. The original document will not be modified.
func InsertMany[T any](ctx context.Context, client *mongo.Database, collection string, data []T) error {
	nCtx, cancel := context.WithTimeout(ctx, queryTimeout*time.Second)
	defer cancel()

	docs := make([]interface{}, len(data))

	for i := range data {
		docs[i] = data[i]
	}

	if _, err := client.Collection(collection).InsertMany(nCtx, docs); err != nil {
		return fmt.Errorf("failed to insert many documents: %v", err)
	}

	return nil
}

// DeleteOne executes a delete command to delete at most one document from the collection.
// If no element was deleted the function will not return an error.
func DeleteOne(ctx context.Context, client *mongo.Database, collection, id string) error {
	nCtx, cancel := context.WithTimeout(ctx, queryTimeout*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}

	if _, err := client.Collection(collection).DeleteOne(nCtx, filter); err != nil {
		return fmt.Errorf("failed to delete document: %v", err)
	}

	return nil
}

// UpdateOne executes an update command to update at most one document in the collection.
// If no element was updated due to not matching the given id the function will not return an error.
func UpdateOne[T any](ctx context.Context, client *mongo.Database, collection, id string, data T) error {
	nCtx, cancel := context.WithTimeout(ctx, queryTimeout*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}

	update := bson.M{"$set": data}

	opt := options.Update().SetUpsert(false)
	if _, err := client.Collection(collection).UpdateOne(nCtx, filter, update, opt); err != nil {
		return fmt.Errorf("failed to update document: %v", err)
	}

	return nil
}

func getCustomTLSConfig(caFilePath string) (*tls.Config, error) {
	tlsConfig := new(tls.Config)
	certs, err := os.ReadFile(fmt.Sprintf(caFilePath))
	if err != nil {
		return tlsConfig, fmt.Errorf("failed to read CA file: %v", err)
	}

	tlsConfig.RootCAs = x509.NewCertPool()

	if ok := tlsConfig.RootCAs.AppendCertsFromPEM(certs); !ok {
		return tlsConfig, fmt.Errorf("failed to parse pem file")
	}

	return tlsConfig, nil
}
