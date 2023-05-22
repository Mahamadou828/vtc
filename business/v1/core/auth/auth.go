package auth

import (
	"context"
	"fmt"
	"time"

	model "vtc/business/v1/data/models/auth"
	"vtc/business/v1/sys/aws/cognito"
	"vtc/business/v1/sys/stripe"
	"vtc/business/v1/sys/validate"
	"vtc/business/v1/web"
)

func SignUp(ctx context.Context, data model.NewUserDTO, cfg *web.AppConfig, agg string, now time.Time) (model.User, error) {
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
