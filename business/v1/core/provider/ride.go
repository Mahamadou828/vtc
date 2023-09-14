package provider

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"vtc/business/v1/data/models"
	"vtc/business/v1/sys/stripe"
	"vtc/foundation/config"
)

// CreatePayment create a new payment for an offer. The created payment is not save in our database upon creation
// but rather when the ride will get booked by the user
func CreatePayment(ctx context.Context, data models.CreatePaymentDTO, cfg *config.App, now time.Time) (stripe.Charge, error) {
	u, err := models.FindOne[models.User](ctx, cfg.DBClient, models.UserCollection, bson.D{{"_id", data.UserID}})
	if err != nil {
		return stripe.Charge{}, fmt.Errorf("failed to find user with id: %v, [%w]", data.UserID, err)
	}

	of, err := models.FindOne[models.Offer](ctx, cfg.DBClient, models.OfferCollection, bson.D{{"_id", data.OfferID}})
	if err != nil {
		return stripe.Charge{}, fmt.Errorf("failed to find offer with id: %v, [%w]", data.OfferID, err)
	}

	var paymentMethod *models.PaymentMethod
	for _, pm := range u.PaymentMethods {
		if pm.IsFavorite {
			paymentMethod = &pm
		}
	}
	if paymentMethod == nil {
		return stripe.Charge{}, fmt.Errorf("user has no valid credit card")
	}

	charge, err := stripe.CreateCharge(cfg.Env.Stripe.Key, of.ProviderPrice, u.StripeID, paymentMethod.StripeID, data.ReturnURL, "eur")
	if err != nil {
		return stripe.Charge{}, fmt.Errorf("failed to create a charge for given payment method and user id: %v, %v", u.StripeID, paymentMethod.ID)
	}

	return charge, nil
}

func RequestRide() {}

func GetRide() {}

func RefreshRide() {}

func CancelRide() {}
