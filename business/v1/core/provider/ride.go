package provider

import (
	"context"
	"fmt"
	"time"
	"vtc/business/v1/sys/provider"
	"vtc/business/v1/sys/validate"

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

func RequestRide(ctx context.Context, data models.NewRideDTO, cfg *config.App, now time.Time) (models.Ride, error) {
	u, err := models.FindOne[models.User](ctx, cfg.DBClient, models.UserCollection, bson.D{{"_id", data.UserID}})
	if err != nil {
		return models.Ride{}, fmt.Errorf("user with id %v not found: %w", data.UserID, err)
	}

	of, err := models.FindOne[models.Offer](ctx, cfg.DBClient, models.OfferCollection, bson.D{{"_id", data.OfferID}})
	if err != nil {
		return models.Ride{}, fmt.Errorf("offer with id %v not found: %w", data.OfferID, err)
	}

	pi, err := stripe.GetPaymentIntent(cfg.Env.Stripe.Key, data.StripeIntentID)
	if err != nil {
		return models.Ride{}, fmt.Errorf("no payment with id %v found: %w", data.StripeIntentID, err)
	}

	if pi.Status != stripe.PaymentIntentStatusSucceeded && pi.Status != stripe.PaymentIntentStatusRequiresCapture {
		return models.Ride{}, fmt.Errorf("3DS process failed, please request a new ride and change the payment_method method")
	}

	rideInfo, err := provider.New(cfg).RequestRide(ctx, *of, provider.UserInfo{ID: u.ID, MySamID: u.MySamClientID}, of.Search, now)
	if err != nil {
		return models.Ride{}, fmt.Errorf("failed to request ride: [%w]", err)
	}

	payment := models.Payment{
		Date:            now,
		Status:          string(pi.Status),
		PreAuthID:       data.StripeIntentID,
		PreAuthPrice:    rideInfo.Price,
		PaymentMethodID: pi.PaymentMethod.ID,
		CreatedAt:       now.String(),
		UpdatedAt:       now.String(),
	}

	ride := models.Ride{
		ID:                  validate.GenerateID(),
		ProviderName:        of.Provider,
		UserID:              u.ID,
		OfferID:             of.ID,
		ProviderRideID:      rideInfo.Id,
		IsPlanned:           of.IsPlanned,
		ETA:                 rideInfo.ETA,
		CancellationFees:    0,
		StartDate:           of.StartDate,
		PaymentByTGS:        true,
		Aggregator:          data.AggregatorCode,
		Status:              rideInfo.Status,
		ProviderPrice:       rideInfo.Price,
		DisplayPrice:        of.DisplayPrice,
		DisplayPriceNumeric: of.DisplayPriceNumeric,
		PriceStatus:         "pending",
		Review:              models.Review{},
		Invoice:             models.Invoice{},
		Payment:             payment,
		Driver:              rideInfo.Driver,
		CreatedAt:           now.String(),
		UpdatedAt:           now.String(),
	}

	if err := models.InsertOne[models.Ride](ctx, cfg.DBClient, models.RideCollection, &ride); err != nil {
		return models.Ride{}, fmt.Errorf("failed to save ride: %v", err)
	}

	return ride, err
}

func GetRide(ctx context.Context, id string, cfg *config.App) (models.Ride, error) {
	ride, err := models.FindOne[models.Ride](ctx, cfg.DBClient, models.RideCollection, bson.D{{"_id", id}})
	return *ride, err
}

func CancelRide() {}

func RefreshRide() {}
