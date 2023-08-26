package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.mongodb.org/mongo-driver/mongo"
	"vtc/business/v1/sys/aws/ssm"
	"vtc/business/v1/sys/database"
)

// Env defines all environment variable needed to run the application
type Env struct {
	Cognito struct {
		ClientID string `conf:"env:COGNITO_CLIENT_ID,required"`
	}
	Stripe struct {
		Key string `conf:"env:STRIPE_KEY,required"`
	}
	Providers struct {
		Timeout int `conf:"env:PROVIDERS_DEFAULT_TIMEOUT"`
		MySam   struct {
			APIKey string `conf:"env:MY_SAM_API_KEY"`
		}
	}
}

var (
	ErrAWSSessionFailure = errors.New("failed to init aws session")
	ErrSecretNotFound    = errors.New("failed to fetch database secret")
)

// App defines all the necessary dependencies to run the application
type App struct {
	DBClient   *mongo.Database
	AWSSession *session.Session
	Env        Env
}

// NewApp create a new App defining all dependencies needed to run the application
func NewApp() (*App, error) {
	//init a new aws session
	sess, err := session.NewSession(
		&aws.Config{
			Region:                        aws.String(os.Getenv("AWS_REGION")),
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
	)
	if err != nil {
		return nil, ErrAWSSessionFailure
	}

	//parsing if ssl is enabled for database connection
	SSLEnable, err := strconv.ParseBool(os.Getenv("DATABASE_SSL_ENABLE"))
	if err != nil {
		return nil, fmt.Errorf("invalid argument for DATABASE_SSL_ENABLE: %v", err)
	}

	//fetch the secret to connect to the database
	secrets, err := ssm.GetSecrets(sess, os.Getenv("DATABASE_POOL_NAME"))
	if err != nil {
		return nil, ErrSecretNotFound
	}

	// open new database client
	client, err := database.NewClient(database.Config{
		Username:   secrets["username"],
		Password:   secrets["password"],
		Host:       secrets["host"],
		Port:       secrets["port"],
		Database:   os.Getenv("DATABASE_NAME"),
		SSLEnabled: SSLEnable,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create a database client: %v", err)
	}

	var env Env
	if err := ParseEnv(&env); err != nil {
		return nil, fmt.Errorf("failed to extract required env config: %v", err)
	}

	return &App{client, sess, env}, nil
}
