package user

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"vtc/business/v1/data/models"
	model "vtc/business/v1/data/models"
	"vtc/business/v1/sys/aws/cognito"
	"vtc/business/v1/sys/stripe"
	"vtc/business/v1/sys/validate"
	"vtc/foundation/config"
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

	if err := models.InsertOne[model.User](ctx, cfg.DBClient, model.UserCollection, &user); err != nil {
		return model.User{}, fmt.Errorf("failed to insert user: %v", err)
	}

	return user, nil
}

// Login log a user and return a new Session
func Login(ctx context.Context, cred model.LoginDTO, cfg *config.App, agg string) (Session, error) {
	// fetch user from database
	u, err := models.FindOne[model.User](ctx, cfg.DBClient, model.UserCollection, bson.D{{"email", cred.Email}, {"aggregator", agg}})
	if err != nil {
		return Session{}, fmt.Errorf("failed to find user: %v, error: %v", cred.Email, err)
	}

	// log the user inside cognito
	tokens, err := cognito.Login(cfg.AWSSession, cfg.Env.Cognito.ClientID, u.CognitoID, cred.Password)
	if err != nil {
		return Session{}, fmt.Errorf("failed to log user: %v, error: %v", cred.Email, err)
	}

	//return a new session
	return Session{User: *u, Tokens: tokens}, nil
}

// CreatePaymentMethod register a new user payment method, if 3DS is needed the URL will be send back with the response
func CreatePaymentMethod(ctx context.Context, data model.NewPaymentMethodDTO, cfg *config.App, now time.Time) (stripe.PaymentIntent, error) {
	u, err := models.FindOne[model.User](ctx, cfg.DBClient, model.UserCollection, bson.D{{"id", data.UserID}})
	if err != nil {
		return stripe.PaymentIntent{}, fmt.Errorf("failed to find user with id: %v", data.UserID)
	}

	pi, err := stripe.RegisterCard(cfg.Env.Stripe.Key, u.StripeID, data)
	if err != nil {
		return stripe.PaymentIntent{}, fmt.Errorf("failed to register a new credit card: [%w]", err)
	}

	// mark all other payment method as non-favorite since we can have only one favorite pm
	if data.IsFavorite {
		for _, pm := range u.PaymentMethods {
			pm.IsFavorite = false
		}
	}

	pm := model.PaymentMethod{
		Name:              data.PaymentMethodName,
		Active:            true,
		CreditCardPayload: data.CardNumber[:3],
		IntentID:          pi.IntentID,
		PaymentServiceID:  pi.PaymentMethodID,
		CreditCardType:    pi.CardType,
		IsFavorite:        data.IsFavorite,
		CreatedAt:         now.String(),
		UpdatedAt:         now.String(),
		DeletedAt:         "",
	}

	u.PaymentMethods = append(u.PaymentMethods, pm)

	if err := models.UpdateOne[model.User](ctx, cfg.DBClient, model.UserCollection, u.ID, u); err != nil {
		return stripe.PaymentIntent{}, fmt.Errorf("failed to update user payment method: [%w]", err)
	}

	return pi, nil
}
