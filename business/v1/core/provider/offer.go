package provider

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	mOffer "vtc/business/v1/data/models/offer"
	mUser "vtc/business/v1/data/models/user"
	"vtc/business/v1/sys/provider"
	"vtc/business/v1/sys/validate"
	"vtc/foundation/config"
)

// GetOffers return all offer that match the given search
func GetOffers(ctx context.Context, data mOffer.GetOfferDTO, cfg *config.App, agg string, now time.Time) ([]mOffer.Offer, error) {
	u, err := mUser.FindOne(ctx, cfg.DBClient, bson.D{{"_id", data.UserID}})
	if err != nil {
		return nil, fmt.Errorf("failed to find user %v: [%w]", data.UserID, err)
	}

	startDate, isPlanned := now, false
	if len(data.StartDate) > 0 {
		date, err := time.Parse(time.RFC3339, data.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %v", err)
		}
		if diff := date.Sub(time.Now()); diff.Hours() < 2 {
			return nil, fmt.Errorf("start date should be 2 hours in advance")
		}

		startDate, isPlanned = date, true
	}

	search := mOffer.Search{
		ID:         validate.GenerateID(),
		UserID:     u.ID,
		Aggregator: agg,

		StartAddress:   data.StartAddress,
		StartLatitude:  data.StartLatitude,
		StartLongitude: data.StartLongitude,
		StartCountry:   data.StartCountry,

		EndAddress:   data.EndAddress,
		EndLatitude:  data.EndLatitude,
		EndLongitude: data.EndLongitude,
		EndCountry:   data.EndCountry,

		StartDate:      startDate,
		AskedProvider:  data.ProviderList,
		Distance:       data.Distance,
		NbrOfPassenger: data.NbrOfPassenger,
		IsPlanned:      isPlanned,

		CreatedAt: now.String(),
		UpdatedAt: now.String(),
		DeletedAt: "",
	}

	offers, err := provider.New(cfg).GetOffers(ctx, provider.UserInfo{ID: u.ID}, search, now)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch offer: [%w]", err)
	}

	if err := mOffer.InsertMany(ctx, cfg.DBClient, offers); err != nil {
		return nil, fmt.Errorf("failed to save offers: [%w]", err)
	}

	return offers, nil
}
