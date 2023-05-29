package cognito_test

import (
	// load env variable from secret manager
	_ "vtc/business/v1/sys/test"

	"log"
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

var (
	sess *session.Session
	u    cognito.User
)

func TestMain(m *testing.M) {
	// init aws session
	var err error
	sess, err = session.NewSession()
	if err != nil {
		log.Fatalf("\t%s\t Test: \tShould be able to open a new session: %v", failure, err)
	}

	//mock a user
	u.PhoneNumber = "+33756866932"
	if err := faker.FakeData(&u); err != nil {
		log.Fatalf("\t%s\t Test: \tShould be able to fake u data: %v", failure, err)
	}

	//Run test and exit
	os.Exit(m.Run())
}

func Test_Signup(t *testing.T) {
	t.Log("Given the need to signup a user")
	{
		if _, err := cognito.SignUp(sess, u, os.Getenv("COGNITO_CLIENT_ID")); err != nil {
			t.Fatalf("\t%s\t Test: \tShould be able to signup a new user: %v", failure, err)
		}
		t.Logf("\t%s\t Test: \tShould be able to signup a new user", success)
	}
}

func Test_Login(t *testing.T) {
	t.Log("Given the need to log a user")
	{
		userID := cognito.GenerateSub(u.Email, u.Password, os.Getenv("COGNITO_CLIENT_ID"))
		if _, err := cognito.Login(sess, os.Getenv("COGNITO_CLIENT_ID"), userID, u.Password); err != nil {
			t.Fatalf("\t%s\t Test: \tShould be able to login the new user: %v", failure, err)
		}
		t.Logf("\t%s\t Test: \tShould be able to login the new user", success)
	}
}
