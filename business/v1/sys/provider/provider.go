package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"vtc/business/v1/data/models"
	"vtc/foundation/config"
)

var (
	ErrFailedToMarshalRequest = errors.New("failed to marshal request body")
	ErrFailedToCreateRequest  = errors.New("failed to create new request")
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
	ID          string
	MySamID     string
	FirstName   string
	LastName    string
	PhoneNumber string
}

// IProvider represent any services that can return and handle ride process
type IProvider interface {
	GetOffers(ctx context.Context, u UserInfo, s models.Search, now time.Time) ([]models.Offer, error)
	RequestRide(ctx context.Context, o models.Offer, u UserInfo, s models.Search, now time.Time) (models.ProviderRide, error)
	GetRide(ctx context.Context, ride models.Ride) (models.ProviderRide, error)
	CancelRide(ctx context.Context, ride models.Ride) (models.ProviderRide, error)
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

func (p Integrations) GetOffers(ctx context.Context, u UserInfo, s models.Search, now time.Time) ([]models.Offer, error) {
	var res []models.Offer

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

func (p Integrations) RequestRide(ctx context.Context, o models.Offer, u UserInfo, s models.Search, now time.Time) (models.ProviderRide, error) {
	ride, err := p.providers[o.Provider].RequestRide(ctx, o, u, s, now)
	if err != nil {
		return models.ProviderRide{}, fmt.Errorf("failed to request ride for the given provider %v: %w", o.Provider, err)
	}

	return ride, err
}

func (p Integrations) CancelRide(ctx context.Context, ride models.Ride) (models.ProviderRide, error) {
	return p.providers[ride.ProviderName].CancelRide(ctx, ride)
}
