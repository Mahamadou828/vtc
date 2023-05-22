// Package lambdasetup create and export an aws session. The package also export the SLLEnable parameter for database connection
// this package should be imported inside all lambda function
package lambdasetup

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"os"
	"strconv"
	"vtc/business/v1/sys/aws/ssm"
	"vtc/business/v1/sys/database"
)

var DatabaseConfig database.Config

func init() {
	ses, err := session.NewSession(&aws.Config{
		Region:                        aws.String(os.Getenv("AWS_REGION")),
		CredentialsChainVerboseErrors: aws.Bool(true),
		Credentials:                   credentials.NewEnvCredentials(),
	})
	if err != nil {
		log.Fatalf("can't init a new aws session: %v", err)
	}

	SSLEnable, err := strconv.ParseBool(os.Getenv("DATABASE_SSL_ENABLE"))
	if err != nil {
		log.Fatalf("invalid argument for DATABASE_SSL_ENABLE: %v", err)
	}

	//fetch the secret to connect to the database
	secrets, err := ssm.GetSecrets(ses, os.Getenv("DATABASE_POOL_NAME"))
	if err != nil {
		log.Fatalf("failed to fetch secret to construct database config: %v", err)
	}

	DatabaseConfig = database.Config{
		Username:   secrets["username"],
		Password:   secrets["password"],
		Host:       secrets["host"],
		Port:       secrets["port"],
		Database:   os.Getenv("DATABASE_NAME"),
		SSLEnabled: SSLEnable,
	}
}
