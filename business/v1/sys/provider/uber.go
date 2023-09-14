// @note the implementation is not completed yet.
package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"vtc/business/v1/data/models"
	"vtc/business/v1/sys/validate"
	"vtc/foundation/config"
)

var (
	UberFailedToFetchApi       = errors.New("failed to fetch uber service")
	UberUnexpectedResponseBody = errors.New("failed to decode uber response body")
)

type Uber struct {
	Client        *http.Client
	OfferMapping  map[string]string
	StatusMapping map[string]string
	LogoURL       string
	Cookie        string
	BaseURL       string
}

type UberResponseOffer struct {
	Status string    `json:"status"`
	Data   UberOffer `json:"data"`
}

type UberResponseRide struct {
	Status string `json:"status"`
	Data   struct {
		Rides []UberRideData `json:"rides"`
	} `json:"data"`
}

type UberOffer struct {
	ProductEstimates []ProductEstimates `json:"productEstimates"`
	FaresUnavailable bool               `json:"faresUnavailable"`
}

type ProductEstimates struct {
	Product  Product  `json:"product"`
	Estimate Estimate `json:"estimate"`
}

type Product struct {
	ProductID         string `json:"productID"`
	DisplayName       string `json:"displayName"`
	Description       string `json:"description"`
	Image             string `json:"image"`
	BackgroundImage   string `json:"backgroundImage"`
	Capacity          int    `json:"capacity"`
	SchedulingEnabled bool   `json:"schedulingEnabled"`
	VVID              int    `json:"vvid"`
	ReserveInfo       struct {
		Enabled                          bool `json:"enabled"`
		ScheduledThresholdMinutes        int  `json:"scheduledThresholdMinutes"`
		FreeCancellationThresholdMinutes int  `json:"freeCancellationThresholdMinutes"`
	} `json:"reserveInfo"`
	ParentProductTypeID            string  `json:"parentProductTypeID"`
	CancellationFee                float64 `json:"cancellationFee"`
	CancellationGracePeriodSeconds int     `json:"cancellationGracePeriodInSeconds"`
}

type Estimate struct {
	FareID                  string          `json:"fareID"`
	PickupEstimateInMinutes int             `json:"pickupEstimateInMinutes"`
	Fare                    Fare            `json:"fare"`
	Trip                    Trip            `json:"trip"`
	PricingExplanation      string          `json:"pricingExplanation"`
	NoCarsAvailable         bool            `json:"noCarsAvailable"`
	Estimate                EstimateDetails `json:"estimate"`
}

type Fare struct {
	Display      string  `json:"display"`
	ExpiresAt    int     `json:"expiresAt"`
	FareID       string  `json:"fareID"`
	CurrencyCode string  `json:"currencyCode"`
	FareValue    float64 `json:"fareValue"`
}

type UberCoordinate struct {
	latitude  float64
	longitude float64
}

type UberAddress struct {
	id           string
	name         string
	addressLine1 string
	addressLine2 string
	fullAddress  string
	coordinate   UberCoordinate
	locale       string
	provider     string
	timeZone     string
}

type EstimateDetails struct {
	CurrencyCode string  `json:"currencyCode"`
	Display      string  `json:"display"`
	HighEstimate float64 `json:"highEstimate"`
	LowEstimate  float64 `json:"lowEstimate"`
}

type Trip struct {
	DistanceUnit              string  `json:"distanceUnit"`
	DistanceEstimate          float64 `json:"distanceEstimate"`
	DurationEstimateInSeconds int     `json:"durationEstimateInSeconds"`
}

type UberRideData struct {
	UUID              string      `json:"uuid"`
	OrganizationUuuid string      `json:"organizationUuuid"`
	Kind              string      `json:"kind"`
	Rider             Ride        `json:"rider"`
	Riders            []Ride      `json:"-"`
	Driver            Driver      `json:"driver"`
	Vehicle           Vehicle     `json:"vehicle"`
	RideDetails       RideDetails `json:"rideDetails"`
	AcceptedAt        string      `json:"acceptedAt"`
}

type Ride struct {
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	FullName    string `json:"fullName,omitempty"`
	Id          string `json:"id,omitempty"`
	Country     struct {
		Label    string `json:"label,omitempty"`
		Id       string `json:"id,omitempty"`
		DialCode string `json:"dialCode,omitempty"`
	} `json:"country"`
	Locale string `json:"locale,omitempty"`
}

type Driver struct {
	Name        string `json:"name,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	Rating      string `json:"rating,omitempty"`
	PictureUrl  string `json:"pictureUrl,omitempty"`
}

type Vehicle struct {
	LicensePlate     string `json:"licensePlate,omitempty"`
	VehicleColorName string `json:"vehicleColorName,omitempty"`
	PictureUrl       string `json:"pictureUrl,omitempty"`
	CarName          string `json:"carName,omitempty"`
}

type RideDetails struct {
	Waypoints            interface{} `json:"-"`
	DurationSeconds      int         `json:"durationSeconds,omitempty"`
	DeferredPickupDay    interface{} `json:"-"`
	CanTip               bool        `json:"canTip,omitempty"`
	DistanceMiles        int         `json:"distanceMiles,omitempty"`
	ClientFare           string      `json:"clientFare,omitempty"`
	PolicyUuid           string      `json:"policyUuid,omitempty"`
	CallEnabled          bool        `json:"callEnabled,omitempty"`
	ExpenseCode          interface{} `json:"-"`
	ExpenseMemo          interface{} `json:"-"`
	NoteForDriver        interface{} `json:"-"`
	RequesterName        string      `json:"requesterName,omitempty"`
	ClientFareNumeric    int         `json:"clientFareNumeric,omitempty"`
	ClientFareWithoutTip string      `json:"clientFareWithoutTip,omitempty"`
	CityID               string      `json:"cityID,omitempty"`
	ReserveDetails       struct {
		IsReserve         bool `json:"isReserve,omitempty"`
		IsReserveAppeased bool `json:"isReserveAppeased,omitempty"`
	} `json:"reserveDetails"`
	TipDetails    interface{}          `json:"-"`
	DropoffTime   interface{}          `json:"-"`
	BeginTripTime interface{}          `json:"-"`
	Pickup        PickupAndDestination `json:"pickup"`
	Product       struct {
		DisplayName         string `json:"displayName,omitempty"`
		ProductID           string `json:"productID,omitempty"`
		ParentProductTypeID string `json:"parentProductTypeID,omitempty"`
	} `json:"product"`
	Status         string `json:"status,omitempty"`
	DriverLocation struct {
		Bearing   int `json:"bearing,omitempty"`
		Latitude  int `json:"latitude,omitempty"`
		Longitude int `json:"longitude,omitempty"`
	} `json:"driverLocation"`
	Destination         PickupAndDestination `json:"destination"`
	RequestTime         int                  `json:"requestTime,omitempty"`
	ScheduledPickupTime string               `json:"scheduledPickupTime,omitempty"`
}

type PickupAndDestination struct {
	Eta       int    `json:"eta,omitempty"`
	Latitude  int    `json:"latitude,omitempty"`
	Longitude int    `json:"longitude,omitempty"`
	Timezone  string `json:"timezone,omitempty"`
	Address   string `json:"address,omitempty"`
	Title     string `json:"title,omitempty"`
	Subtitle  string `json:"subtitle,omitempty"`
	ID        string `json:"id,omitempty"`
	Provider  string `json:"provider,omitempty"`
}

func NewUber(client *http.Client, cfg *config.App) Uber {
	return Uber{
		Client:  client,
		BaseURL: "https://central.uber.com/v2/api",
		Cookie:  cfg.Env.Providers.Uber.Cookie,
		LogoURL: "https://helios-i.mashable.com/imagery/articles/03y6VwlrZqnsuvnwR8CtGAL/hero-image.fill.size_1200x675.v1623372852.jpg",

		OfferMapping:  map[string]string{},
		StatusMapping: map[string]string{},
	}
}

func (u Uber) GetOffers(ctx context.Context, _ UserInfo, s models.Search, now time.Time) ([]models.Offer, error) {

	reqBody := struct {
		pickup           UberAddress
		dropoff          UberAddress
		capacity         int
		scheduling       int64
		rideSessionUuid  string
		organizationUuid string
	}{
		pickup: UberAddress{
			id:           "EiZQbC4gZGUgbCdFcXVlcnJlLCA4MzAwMCBUb3Vsb24sIEZyYW5jZSIuKiwKFAoSCSVBdgYSG8kSEVOj8oaQ_JWqEhQKEgn9sjV7AhvJEhEwyI_9pRkIBA",
			name:         s.StartAddress,
			addressLine1: s.StartAddress,
			addressLine2: s.StartAddress,
			fullAddress:  s.StartAddress,
			coordinate:   UberCoordinate{s.StartLatitude, s.StartLongitude},
			locale:       s.StartCountry,
			provider:     "google_places",
			timeZone:     "Europe/Paris",
		},
		dropoff: UberAddress{
			id:           "ChIJg-EqoxobyRIRJtIEOeIpulo",
			name:         s.EndAddress,
			addressLine1: s.EndAddress,
			addressLine2: s.EndAddress,
			fullAddress:  s.EndAddress,
			coordinate:   UberCoordinate{s.EndLatitude, s.EndLongitude},
			locale:       s.EndCountry,
			provider:     "google_places",
			timeZone:     "Europe/Paris",
		},
		capacity:         1,
		rideSessionUuid:  "1bc7f592-711a-4079-ac98-342eefd73e1f",
		organizationUuid: "ee840421-c340-5053-b46a-37914dd7224d",
	}

	if s.IsPlanned {
		milliseconds := s.StartDate.UTC().UnixNano() / int64(time.Millisecond)
		reqBody.scheduling = milliseconds
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToMarshalRequest, err)
	}

	req, err := u.newRequest(ctx, http.MethodPost, u.endpoint(fmt.Sprintf("getProductEstimates?localeCode=%v", s.StartCountry)), bytes.NewReader(data))

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToCreateRequest, err)
	}

	resp, err := u.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", UberFailedToFetchApi, err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusBadRequest:
		var data any
		json.NewDecoder(resp.Body).Decode(&data)
		return nil, fmt.Errorf("failed to fetch offer for uber, receive following error: %v", data)
	case http.StatusOK:
		var data UberResponseOffer

		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&data); err != nil {
			return nil, fmt.Errorf("%w: %v", UberUnexpectedResponseBody, err)
		}

		var filteredUberOffers []ProductEstimates

		for _, offer := range data.Data.ProductEstimates {
			if _, ok := u.OfferMapping[offer.Product.DisplayName]; !ok {
				continue
			}
			filteredUberOffers = append(filteredUberOffers, offer)
		}

		var res []models.Offer
		for _, offer := range filteredUberOffers {
			res = append(res, u.convertProviderOffer(offer, s, now))
		}
		return res, nil
	default:
		var data any
		json.NewDecoder(resp.Body).Decode(&data)
		return nil, fmt.Errorf("unsupported status response, receive status %v and response body %v", resp.StatusCode, data)

	}

}

func (u Uber) RequestRide(ctx context.Context, o models.Offer, ui UserInfo, s models.Search, now time.Time) (models.ProviderRide, error) {
	offerMetadata, err := url.ParseQuery(o.ProviderOfferID)
	if err != nil {
		return models.ProviderRide{}, fmt.Errorf("missing metadata inside provider offer id: [%w]", err)
	}

	ubervvid, _ := strconv.Atoi(offerMetadata.Get("ubervvid"))

	type guest struct {
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		PhoneNumber string `json:"phoneNumber"`
		PhonePrefix string `json:"phonePrefix"`
		Locale      string `json:"locale"`
	}

	type estimate struct {
		FareID                  string          `json:"fareID"`
		Fare                    Fare            `json:"fare"`
		PickupEstimateInMinutes int             `json:"pickupEstimateInMinutes"`
		Trip                    Trip            `json:"trip"`
		PricingExplanation      string          `json:"pricingExplanation"`
		NoCarsAvailable         bool            `json:"noCarsAvailable"`
		Estimate                EstimateDetails `json:"estimate"`
	}

	type tripLeg struct {
		AdditionalStops []interface{} `json:"-"`
		Capacity        int           `json:"capacity"`
		ExpenseMemo     string        `json:"expenseMemo"`
		NoteForDriver   string        `json:"noteForDriver"`
		PickupAddress   UberAddress   `json:"pickupAddress"`
		DropoffAddress  UberAddress   `json:"dropoffAddress"`
		Product         Product       `json:"product"`
		Estimate        estimate      `json:"estimate"`
		Scheduling      struct {
			PickupTime int64 `json:"pickupTime"`
		} `json:"scheduling"`
	}

	reqBody := struct {
		Guest                guest         `json:"guest"`
		AdditionalGuests     []interface{} `json:"additionalGuests"`
		TripLegs             []tripLeg     `json:"tripLegs"`
		ExpenseCode          string        `json:"expenseCode"`
		BypassSmsOptOutCheck bool          `json:"bypassSmsOptOutCheck"`
		CallEnabled          bool          `json:"callEnabled"`
		PolicyUuid           string        `json:"policyUuid"`
		RideSessionUuid      string        `json:"rideSessionUuid"`
		OrganizationUuid     string        `json:"organizationUuid"`
	}{
		Guest: guest{
			FirstName:   ui.FirstName,
			LastName:    ui.LastName,
			PhoneNumber: ui.PhoneNumber,
			Locale:      s.StartCountry,
		},
		AdditionalGuests: nil,
		TripLegs: []tripLeg{
			{AdditionalStops: nil,
				Capacity:      1,
				ExpenseMemo:   "",
				NoteForDriver: s.StartAddress,
				PickupAddress: UberAddress{
					id:           "de93a71f-7782-4894-a114-5e357de81fa1",
					name:         s.StartAddress,
					addressLine1: s.StartAddress,
					addressLine2: s.StartAddress,
					fullAddress:  s.StartAddress,
					coordinate:   UberCoordinate{s.StartLatitude, s.StartLongitude},
					locale:       s.StartCountry,
					provider:     "uber_geofences",
					timeZone:     "Europe/Paris",
				},
				DropoffAddress: UberAddress{
					id:           "EhxSdWUgZGUgUml2b2xpLCBQYXJpcywgRnJhbmNlIi4qLAoUChIJt4MohSFu5kcRUHvqO0vC-IgSFAoSCQ-34gYfbuZHEWCUjGjDggsE",
					name:         s.EndAddress,
					addressLine1: s.EndAddress,
					addressLine2: s.EndAddress,
					fullAddress:  s.EndAddress,
					coordinate:   UberCoordinate{s.EndLatitude, s.EndLongitude},
					locale:       s.EndCountry,
					provider:     "uber_geofences",
					timeZone:     "Europe/Paris",
				},
				Product: Product{
					ProductID:         offerMetadata.Get("productID"),
					DisplayName:       o.ProviderOfferName,
					Description:       o.Description,
					Image:             "",
					BackgroundImage:   "",
					Capacity:          3,
					SchedulingEnabled: true,
					VVID:              ubervvid,
					ReserveInfo: struct {
						Enabled                          bool `json:"enabled"`
						ScheduledThresholdMinutes        int  `json:"scheduledThresholdMinutes"`
						FreeCancellationThresholdMinutes int  `json:"freeCancellationThresholdMinutes"`
					}{
						Enabled:                          false,
						ScheduledThresholdMinutes:        120,
						FreeCancellationThresholdMinutes: 60,
					},
					ParentProductTypeID: "6a8e56b8-914e-4b48-a387-e6ad21d9c00c",
				},
				Estimate: estimate{
					FareID: offerMetadata.Get("uberFareID"),
					Fare: Fare{
						Display:      fmt.Sprintf("%v %v", o.ProviderPrice, '€'),
						ExpiresAt:    1636890197,
						FareID:       offerMetadata.Get("uberFareID"),
						CurrencyCode: "eur",
					},
					PickupEstimateInMinutes: 0,
					Trip: Trip{
						DistanceUnit:              "km",
						DistanceEstimate:          1.79,
						DurationEstimateInSeconds: 1050,
					},
					PricingExplanation: "",
					NoCarsAvailable:    false,
				},
			},
		},
		ExpenseCode:          "",
		BypassSmsOptOutCheck: false,
		CallEnabled:          true,
		PolicyUuid:           "f9479684-539c-4480-bae7-b6ff5bdd64e9",
		RideSessionUuid:      "e6fd2463-ad1f-493f-9e6f-dc67648154e7",
		OrganizationUuid:     "ee840421-c340-5053-b46a-37914dd7224d",
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return models.ProviderRide{}, fmt.Errorf("%w: %v", ErrFailedToMarshalRequest, err)
	}

	req, err := u.newRequest(ctx, http.MethodPost, fmt.Sprintf("createRide?localeCode=%v", s.StartCountry), bytes.NewReader(data))
	if err != nil {
		return models.ProviderRide{}, fmt.Errorf("%w: %v", ErrFailedToCreateRequest, err)
	}

	resp, err := u.Client.Do(req)
	if err != nil {
		return models.ProviderRide{}, fmt.Errorf("%w: %v", UberFailedToFetchApi, err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var ride UberResponseRide
		decoder := json.NewDecoder(resp.Body)

		if err := decoder.Decode(&ride); err != nil {
			return models.ProviderRide{}, fmt.Errorf("%w: %v", MySAMUnexpectedResponseBody, err)
		}

		return u.convertProviderRide(ride.Data.Rides[0], o, now), nil
	case http.StatusBadRequest:
		var data any
		json.NewDecoder(resp.Body).Decode(&data)
		return models.ProviderRide{}, fmt.Errorf("failed to fetch offer for mysam, receive following error: %v", data)
	default:
		var data any
		json.NewDecoder(resp.Body).Decode(&data)
		return models.ProviderRide{}, fmt.Errorf("unsupported status response, receive status %v and response body %v", resp.StatusCode, data)
	}
}

func (u Uber) GetRide(ctx context.Context, ride models.Ride) (models.ProviderRide, error) {
	return models.ProviderRide{}, nil
}

func (u Uber) CancelRide(ctx context.Context, ride models.Ride) (models.ProviderRide, error) {
	rideMetadata, err := url.ParseQuery(ride.ProviderRideID)
	if err != nil {
		return models.ProviderRide{}, err
	}

	reqBody := struct {
		RideUUID string `json:"rideUUID"`
	}{
		RideUUID: rideMetadata.Get("uuid"),
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return models.ProviderRide{}, fmt.Errorf("%w: %v", ErrFailedToMarshalRequest, err)
	}

	req, err := u.newRequest(ctx, http.MethodPost, "cancelRide", bytes.NewReader(data))
	if err != nil {
		return models.ProviderRide{}, fmt.Errorf("%w: %v", ErrFailedToCreateRequest, err)
	}

	resp, err := u.Client.Do(req)
	if err != nil {
		return models.ProviderRide{}, fmt.Errorf("%w: %v", UberFailedToFetchApi, err)
	}
	defer resp.Body.Close()

	return models.ProviderRide{
		Id:         ride.ProviderRideID,
		Status:     Cancelled,
		StatusName: "cancelled",
		Price:      0,
		ETA:        0,
		Driver:     models.Driver{},
	}, nil
}

func (u Uber) convertProviderOffer(offer ProductEstimates, s models.Search, now time.Time) models.Offer {
	providerID := make(url.Values)

	providerID.Set("uberFareID", offer.Estimate.FareID)
	providerID.Set("parentProductTypeID", offer.Product.ParentProductTypeID)
	providerID.Set("ubervvid", string(rune(offer.Product.VVID)))
	providerID.Set("productID", offer.Product.ProductID)
	providerID.Set("cancellationFee", fmt.Sprint(offer.Product.CancellationFee))
	providerID.Set("cancellationGracePeriodInSeconds", string(rune(offer.Product.CancellationGracePeriodSeconds)))

	return models.Offer{
		ID:                  validate.GenerateID(),
		StartDate:           s.StartDate.String(),
		Provider:            "uber",
		ETA:                 float64(offer.Estimate.PickupEstimateInMinutes * 60),
		ProviderOfferID:     providerID.Encode(),
		LogoURL:             u.LogoURL,
		VehicleType:         u.OfferMapping[offer.Product.DisplayName],
		ProviderOfferName:   offer.Product.DisplayName,
		ProviderPrice:       offer.Estimate.Fare.FareValue,
		DisplayPrice:        fmt.Sprintf("%f %s", offer.Estimate.Fare.FareValue, "€"),
		DisplayPriceNumeric: offer.Estimate.Fare.FareValue,
		DisplayProviderName: "Uber",
		IsPlanned:           s.IsPlanned,
		UserID:              s.UserID,
		Search:              s,
		Aggregator:          s.Aggregator,
		Description:         "",
		CreatedAt:           now.String(),
		UpdatedAt:           now.String(),
	}
}

func (u Uber) convertProviderRide(ride UberRideData, o models.Offer, now time.Time) models.ProviderRide {
	eta := float64(ride.RideDetails.Pickup.Eta)

	providerID := make(url.Values)

	providerID.Set("uuid", ride.UUID)
	providerID.Set("acceptedAt", ride.AcceptedAt)

	price := float64(ride.RideDetails.ClientFareNumeric)

	if ride.RideDetails.ClientFareNumeric == 0 {
		price = o.ProviderPrice
	}

	status := u.StatusMapping[ride.RideDetails.Status]

	if status == Arriving {
		eta = 0
	}

	return models.ProviderRide{
		Id:         providerID.Encode(),
		Status:     status,
		StatusName: ride.RideDetails.Status,
		Price:      price,
		ETA:        eta,
		Driver:     models.Driver{},
	}
}

func (u Uber) endpoint(path string) string {
	return fmt.Sprintf("%v/%v", u.BaseURL, path)
}

func (u Uber) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header = map[string][]string{
		"authority":                       {"central.uber.com"},
		"sec-ch-ua":                       {"\"Not;A Brand\";v=\"99\", \"Google Chrome\";v=\"97\", \"Chromium\";v=\"97'\""},
		"x-csrf-token":                    {"x"},
		"sec-ch-ua-mobile":                {"?0"},
		"user-agent":                      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36"},
		"x-guest-rides-organization-uuid": {"ee840421-c340-5053-b46a-37914dd7224d"},
		"x-requested-with":                {"XMLHttpRequest"},
		"x-guest-rides-app-version":       {"1.0.0"},
		"content-type":                    {"application/json"},
		"sec-ch-ua-platform":              {"'Windows'"},
		"accept":                          {"*/*"},
		"origin":                          {"https://central.uber.com"},
		"sec-fetch-site":                  {"same-origin"},
		"sec-fetch-mode":                  {"cors"},
		"sec-fetch-dest":                  {"empty"},
		"referer":                         {"https://central.uber.com/v2/organization/ee840421-c340-5053-b46a-37914dd7224d/new-ride?state=xPcpCBbalkSIEaUDTr6ru4DpKSpKOp1wz1a9D2VbSow%3D&_csid=_MCsj2lPFfUTXmMo9SMIkg"},
		"accept-language":                 {"fr,fr-FR;q=0.9,en-US;q=0.8,en;q=0.7,ru;q=0.6,nl;q=0.5"},
		"cookie":                          {u.Cookie},
	}

	return req, nil
}
