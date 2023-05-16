package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"vtc/business/v1/sys/aws/ssm"
)

const (
	caFilePath = "business/v1/sys/database/rds-combined-ca-bundle.pem"

	connectTimeout = 5
	queryTimeout   = 30

	//connectionStringTemplate accepts 3 parameters: username, password, host, port and ssl
	connectionStringTemplate = "mongodb://%s:%s@%s:%s/"
)

// NewClient create a new database client connection. It accepts an aws session and a secret manager pool name to fetch
// the credentials to connect to the database.
func NewClient(ses *session.Session, secretPoolName string, sslEnabled bool) (*mongo.Client, error) {
	//fetch the secret to connect to the database
	secrets, err := ssm.GetSecrets(ses, secretPoolName)
	if err != nil {
		return nil, fmt.Errorf("can't fetch secrets to open db connection: %v", err)
	}

	connectionURI := fmt.Sprintf(connectionStringTemplate, secrets["username"], secrets["password"], secrets["host"], secrets["port"])
	print(connectionURI, "\n")
	tlsConfig, err := getCustomTLSConfig(caFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get TLS configuration: %v", err)
	}

	clientOpt := options.Client().ApplyURI(connectionURI)
	if sslEnabled {
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

	return client, nil
}

func Find(client *mongo.Client) {

}

func FindOne(client *mongo.Client) {

}

func InsertOne(client *mongo.Client) {

}

func InsertMany(client *mongo.Client) {

}

func Delete(client *mongo.Client) {

}

func Update(client *mongo.Client) {

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
