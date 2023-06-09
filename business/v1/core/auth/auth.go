package auth

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"time"
	"vtc/foundation/config"

	model "vtc/business/v1/data/models/auth"
	"vtc/business/v1/sys/aws/cognito"
	"vtc/business/v1/sys/stripe"
	"vtc/business/v1/sys/validate"
)

type Session struct {
	User   model.User      `json:"user"`
	Tokens cognito.Session `json:"tokens"`
}

// SignUp create a new user account
func SignUp(ctx context.Context, data model.NewUserDTO, cfg *config.App, agg string, now time.Time) (model.User, error) {
	// create the user in cognito pool
	id, err := cognito.SignUp(
		cfg.AWSSession,
		cognito.User{
			Email:       data.Email,
			PhoneNumber: data.PhoneNumber,
			Name:        data.Name,
			Password:    data.Password,
		},
		cfg.Env.Cognito.ClientID,
	)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to create user in cognito pool: %v", err)
	}

	// create stripe account
	stripeID, err := stripe.CreateCustomer(cfg.Env.Stripe.Key, stripe.Customer{
		Email:       data.Email,
		PhoneNumber: data.PhoneNumber,
		Aggregator:  agg,
		Name:        data.Name,
	})
	if err != nil {
		return model.User{}, fmt.Errorf("failed to create stripe user: %v", err)
	}

	//save user in database
	user := model.User{
		ID:               validate.GenerateID(),
		Email:            data.Email,
		PhoneNumber:      data.PhoneNumber,
		Name:             data.Name,
		StripeID:         stripeID,
		MySamClientID:    "",
		Aggregator:       agg,
		PushSubscription: "",
		CognitoID:        id,
		Addresses:        []model.Address{},
		PaymentMethods:   []model.PaymentMethod{},
		CreatedAt:        now.String(),
		UpdatedAt:        "",
		DeletedAt:        "",
	}

	if err := model.InsertOne(ctx, cfg.DBClient, user); err != nil {
		return model.User{}, fmt.Errorf("failed to insert user: %v", err)
	}

	return user, nil
}

// Login log a user and return a new Session
func Login(ctx context.Context, cred model.LoginDTO, cfg *config.App, agg string) (Session, error) {
	// fetch user from database
	u, err := model.FindOne(ctx, cfg.DBClient, bson.D{{"email", cred.Email}, {"aggregator", agg}})
	if err != nil {
		return Session{}, fmt.Errorf("failed to find user: %v, error: %v", cred.Email, err)
	}

	// log the user inside cognito
	tokens, err := cognito.Login(cfg.AWSSession, cfg.Env.Cognito.ClientID, u.CognitoID, cred.Password)
	if err != nil {
		return Session{}, fmt.Errorf("failed to log user: %v, error: %v", cred.Email, err)
	}

	//return a new session
	return Session{u, tokens}, nil
}
