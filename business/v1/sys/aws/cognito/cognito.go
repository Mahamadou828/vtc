package cognito

import (
	"crypto/sha256"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

// User represents all the user data store in cognito
type User struct {
	Email       string
	PhoneNumber string
	Name        string
	Password    string
}

// Session represent a user session obtained after authentication
type Session struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

// SignUp create a new user inside the aws cognito pool. The userID will be a hash of his email, phone number
// and the clientID. This is done to allow multiple email or phone numbers accounts
func SignUp(u User, sess *session.Session, clientID string) (string, error) {
	client := cognitoidentityprovider.New(sess)

	sub := generateSub(u.Email, u.Password, clientID)

	inp := &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(clientID),
		Password: aws.String(u.Password),
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(u.Email),
			},
			{
				Name:  aws.String("phoneNumber"),
				Value: aws.String(u.PhoneNumber),
			},
			{
				Name:  aws.String("name"),
				Value: aws.String(u.Name),
			},
		},
		Username: aws.String(sub),
	}

	if _, err := client.SignUp(inp); err != nil {
		return "", fmt.Errorf("failed to sign up: %v", err)
	}

	return sub, nil
}

// Login create a new access session for the given user
func Login(sess *session.Session, clientID, userID, password string) (Session, error) {
	client := cognitoidentityprovider.New(sess)

	inp := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: nil,
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(userID),
			"PASSWORD": aws.String(password),
			"SRP_A":    aws.String(password),
		},
		ClientId:        aws.String(clientID),
		ClientMetadata:  nil,
		UserContextData: nil,
	}

	res, err := client.InitiateAuth(inp)
	if err != nil {
		return Session{}, fmt.Errorf("failed to log the user: %s", err)
	}

	return Session{
		Token:        *res.AuthenticationResult.AccessToken,
		RefreshToken: *res.AuthenticationResult.RefreshToken,
		ExpiresIn:    int(*res.AuthenticationResult.ExpiresIn),
	}, nil
}

// generateSub create a unique hash from the email, phone number and clientID that will be used as the user's id
func generateSub(email, phoneNumber, clientID string) string {
	sub := email + phoneNumber + clientID
	h := sha256.New()
	h.Write([]byte(sub))
	return fmt.Sprintf("%x", h.Sum(nil))
}
