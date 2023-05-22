package test

import (
	"log"
	"os"
	"vtc/business/v1/sys/aws/ssm"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func init() {
	ses, err := session.NewSession(&aws.Config{
		Region:                        aws.String(os.Getenv("AWS_REGION")),
		CredentialsChainVerboseErrors: aws.Bool(true),
	})
	if err != nil {
		log.Fatalf("can't init a new aws session: %v", err)
	}

	secrets, err := ssm.GetSecrets(ses, "tgs-with-go-db-secret-local")
	if err != nil {
		log.Fatalf("can't fetch env vars for test: %v", err)
	}

	for name, secret := range secrets {
		os.Setenv(name, secret)
	}
}
