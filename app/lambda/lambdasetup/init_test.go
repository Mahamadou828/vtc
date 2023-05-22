package lambdasetup_test

import (
	"fmt"
	"os"
	"testing"
	"vtc/app/lambda/lambdasetup"
	"vtc/business/v1/sys/database"
)

const (
	success = "\u2713"
	failure = "\u2717"
)

func init() {
	os.Setenv("AWS_REGION", "eu-west-1")
	os.Setenv("DATABASE_SSL_ENABLE", "true")
	os.Setenv("DATABASE_POOL_NAME", "tgs-with-go-db-secret-local")
	os.Setenv("DATABASE_NAME", "thegoodseat")
}

func TestInitLambda(t *testing.T) {
	t.Log("Given the need to init config before lambda start")
	{
		t.Log("Given the need to init database config")
		{
			res := database.Config{
				Username:   "user",
				Password:   "password",
				Host:       "host",
				Port:       "27017",
				Database:   "thegoodseat",
				SSLEnabled: true,
			}

			if res != lambdasetup.DatabaseConfig {
				t.Fatalf("\t%s\t Test: \tShould be able to init database config: %v", failure, fmt.Errorf("database config differ from res"))
			}
			t.Logf("\t%s\t Test: \tShould be able to init database config", success)
		}
	}
}
