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
)

var err error
var SES *session.Session
var SSLEnable bool

func init() {
	SES, err = session.NewSession(&aws.Config{
		Region:                        aws.String(os.Getenv("AWS_REGION")),
		CredentialsChainVerboseErrors: aws.Bool(true),
		Credentials:                   credentials.NewEnvCredentials(),
	})
	if err != nil {
		log.Fatalf("can't init a new aws session: %v", err)
	}

	SSLEnable, err = strconv.ParseBool(os.Getenv("DATABASE_SSL_ENABLE"))
	if err != nil {
		log.Fatalf("invalid argument for DATABASE_SSL_ENABLE: %v", err)
	}
}
