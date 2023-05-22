package ssm_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"testing"
	"vtc/business/v1/sys/aws/ssm"
)

const (
	success = "\u2713"
	failure = "\u2717"
)

func TestSSM(t *testing.T) {
	t.Log("Given the need to handle secrets with aws ssm")
	{
		sess, err := session.NewSession()
		if err != nil {
			t.Fatalf("\t%s\t Test: \tShould be able to open a new session: %v", failure, err)
		}

		t.Log("Given the need to fetch secrets")
		{
			secrets, err := ssm.GetSecrets(sess, "tgs-with-go-db-secret-local")
			if err != nil {
				t.Fatalf("\t%s\t Test: \tShould be able to fetch secrets: %v", failure, err)
			}

			if val, ok := secrets["test"]; !ok || val != "test" {
				t.Fatalf("\t%s\t Test: \tShould be able to fetch secrets: %v", failure, fmt.Errorf("secrets test missing value"))
			}

			t.Logf("\t%s\t Test: \tShould be able to fetch secrets", success)
		}
	}
}
