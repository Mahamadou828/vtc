package provider

import (
	"context"
	"net/http"
	"sync"
	"time"
	mRide "vtc/business/v1/data/models/ride"

	mOffer "vtc/business/v1/data/models/offer"
	"vtc/foundation/config"
)

const (
	ECO       = "eco"
	VAN       = "van"
	TwoWheels = "2-wheels"
	EScooter  = "e-scooter"
	Green     = "green"
	Shared    = "shared"
	Business  = "business"
	Access    = "access"
)

const (
	Processing       = "processing"
	Accepted         = "accepted"
	InProgress       = "in_progress"
	Completed        = "completed"
	Arriving         = "arriving"
	Cancelled        = "cancelled"
	DriverCancelled  = "driver_cancelled"
	NoDriverFound    = "no_driver_found"
	Scheduled        = "scheduled"
	OnboardCancelled = "onboard_cancelled"
)

// UserInfo represent all the info needed to get offer and request ride
type UserInfo struct {
	ID      string
	MySamID string
}

// IProvider represent any services that can return and handle ride process
type IProvider interface {
	GetOffers(ctx context.Context, u UserInfo, s mOffer.Search, now time.Time) ([]mOffer.Offer, error)
	RequestRide(ctx context.Context, o mOffer.Offer, u UserInfo, s mOffer.Search, now time.Time) (mRide.Ride, error)
	GetRide(ctx context.Context, ride mRide.Ride) (mRide.Info, error)
	CancelRide(ctx context.Context, ride mRide.Ride) (mRide.Info, error)
	GetCancellationFees()
}

type Integrations struct {
	providers map[string]IProvider
}

func New(cfg *config.App) Integrations {
	client := &http.Client{Timeout: time.Duration(cfg.Env.Providers.Timeout) * time.Second}
	return Integrations{
		providers: map[string]IProvider{
			"mysam": NewMySam(client, cfg),
		},
	}
}

func (p Integrations) GetOffers(ctx context.Context, u UserInfo, s mOffer.Search, now time.Time) ([]mOffer.Offer, error) {
	var res []mOffer.Offer

	var wg sync.WaitGroup
	wg.Add(len(s.AskedProvider))

	var mu sync.Mutex

	for _, provider := range s.AskedProvider {
		go func(provider string) {
			integrations, ok := p.providers[provider]
			// check if the asked provider exist
			if !ok {
				//@todo handle when a non existing provider was query
				wg.Done()
				return
			}

			//fetch offer from the given provider
			offers, err := integrations.GetOffers(ctx, u, s, now)
			if err != nil {
				//@todo handle when provider return an error
				wg.Done()
				return
			}

			// push all offers into result array
			mu.Lock()
			{
				res = append(res, offers...)
			}
			mu.Unlock()

			// finish the given task
			wg.Done()
		}(provider)
	}

	wg.Wait()
	return res, nil
}
