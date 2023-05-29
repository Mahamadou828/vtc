package config_test

import (
	"errors"
	"fmt"
	"os"

	_ "vtc/business/v1/sys/test"
	"vtc/foundation/config"

	"testing"
)

const (
	success = "\u2713"
	failure = "\u2717"
)

func Test_NewApp(t *testing.T) {
	t.Log("Given the need to create an app config")
	{
		t.Log("When handling valid configuration")
		{
			app, err := config.NewApp()
			if err != nil {
				t.Fatalf("\t%s\t Test: \tShould be able to create new app config: %v", failure, err)
			}

			if app.AWSSession == nil {
				t.Logf("\t%s\t Test: \tGiven aws session is invalid : %v", failure, fmt.Errorf("aws session is nil"))
			}

			if *app.AWSSession.Config.Region != "eu-west-1" {
				t.Logf("\t%s\t Test: \tGiven aws region is invalid : %v", failure, fmt.Errorf("aws region is invalid, should be %v, but receive: %v", "eu-west-1", *app.AWSSession.Config.Region))
			}

			if app.DBClient == nil {
				t.Logf("\t%s\t Test: \tGiven db client is invalid : %v", failure, fmt.Errorf("db client is nil"))
			}

			t.Logf("\t%s\t Test: \tShould be able to create new app config:", success)
		}

		t.Log("When handling invalid configuration")
		{
			//changing aws region to test error handling for session
			os.Setenv("AWS_REGION", "xxx")
			_, err := config.NewApp()

			if err == nil || !errors.Is(err, config.ErrAWSSessionFailure) {
				t.Fatalf("\t%s\t Test: \tReceive unexpected error, waiting for: %v but receive : %v", failure, config.ErrAWSSessionFailure, err)
			}

			//changing pool name to test error handling for ssm
			os.Setenv("AWS_REGION", "eu-west-1")
			os.Setenv("DATABASE_POOL_NAME", "xxx")

			_, err = config.NewApp()

			if err == nil || !errors.Is(err, config.ErrSecretNotFound) {
				t.Fatalf("\t%s\t Test: \tReceive unexpected erro, waiting for: %v but receive : %v", failure, config.ErrSecretNotFound, err)
			}
		}
	}
}
