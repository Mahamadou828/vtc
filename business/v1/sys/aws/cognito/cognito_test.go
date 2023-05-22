package cognito_test

import (
	_ "vtc/business/v1/sys/test"

	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-faker/faker/v4"
	"vtc/business/v1/sys/aws/cognito"
)

const (
	success = "\u2713"
	failure = "\u2717"
)

func TestCognito(t *testing.T) {
	t.Log("Given the need to interact with aws cognito")
	{
		sess, err := session.NewSession()
		if err != nil {
			t.Fatalf("\t%s\t Test: \tShould be able to open a new session: %v", failure, err)
		}

		clientID := os.Getenv("COGNITO_CLIENT_ID")
		if len(clientID) == 0 {
			t.Fatalf("\t%s\t Test: \tA Cognito client id should be provide to run test", failure)
		}

		var u cognito.User

		u.PhoneNumber = "+33756866932"

		if err := faker.FakeData(&u); err != nil {
			t.Fatalf("\t%s\t Test: \tShould be able to fake u data: %v", failure, err)
		}

		t.Log("Given the need to signup a user")
		{
			if _, err := cognito.SignUp(sess, u, clientID); err != nil {
				t.Fatalf("\t%s\t Test: \tShould be able to signup a new user: %v", failure, err)
			}
			t.Logf("\t%s\t Test: \tShould be able to signup a new user", success)
		}

		t.Log("Given the need to log a user")
		{
			userID := cognito.GenerateSub(u.Email, u.Password, clientID)
			if _, err := cognito.Login(sess, clientID, userID, u.Password); err != nil {
				t.Fatalf("\t%s\t Test: \tShould be able to login the new user: %v", failure, err)
			}
			t.Logf("\t%s\t Test: \tShould be able to login the new user", success)
		}
	}
}
