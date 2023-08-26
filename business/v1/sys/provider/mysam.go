package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	mOffer "vtc/business/v1/data/models/offer"
	mRide "vtc/business/v1/data/models/ride"
	"vtc/business/v1/sys/validate"
	"vtc/foundation/config"
)

type MySam struct {
	Client        *http.Client
	BaseURL       string
	APIKey        string
	OfferMapping  map[string]string
	StatusMapping map[string]string
	LogoURL       string
}

type MySamRide struct {
	Id             int          `json:"id"`
	FromAddress    MySamAddress `json:"fromAddress"`
	ToAddress      MySamAddress `json:"toAddress"`
	Status         string       `json:"status"`
	StartDate      int64        `json:"startDate"`
	EstimatedPrice float64      `json:"estimatedPrice"`
	FinalPrice     float64      `json:"finalPrice"`
	Driver         *struct {
		FirstName         string `json:"firstName,omitempty"`
		LastName          string `json:"lastName,omitempty"`
		MobilePhoneNumber string `json:"mobilePhoneNumber,omitempty"`
		DriverDetails     *struct {
			VehicleModel string `json:"vehicleModel,omitempty"`
		} `json:"driverDetails,omitempty"`
		Location *struct {
			Latitude  float64 `json:"latitude,omitempty"`
			Longitude float64 `json:"longitude,omitempty"`
		} `json:"location,omitempty"`
	} `json:"driver,omitempty"`
}

type MySamOffer struct {
	Estimation struct {
		StartDate      int64   `json:"startDate"`
		TripType       string  `json:"tripType"`
		VehicleType    string  `json:"vehicleType"`
		Id             int     `json:"id"`
		Created        int64   `json:"created"`
		Distance       int     `json:"distance"`
		Duration       float64 `json:"duration"`
		Increase       int     `json:"increase"`
		Price          float64 `json:"price"`
		PriceIncreased bool    `json:"priceIncreased"`
	} `json:"estimation"`
}

type MySamAddress struct {
	Address   string  `json:"address"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func NewMySam(client *http.Client, cfg *config.App) MySam {
	return MySam{
		Client:  client,
		BaseURL: "https://api.demo.mysam.fr/api",
		APIKey:  cfg.Env.Providers.MySam.APIKey,
		LogoURL: "https://mysam.fr/wp-content/uploads/2019/06/LOGO_MYSAM.png",

		OfferMapping: map[string]string{
			"CAR":   ECO,
			"VAN":   VAN,
			"LUXE":  Business,
			"PRIME": Business,
		},

		StatusMapping: map[string]string{
			"WAITING":             Processing,
			"ASSIGNED":            Accepted,
			"STARTED":             InProgress,
			"FINISHED":            Completed,
			"CANCELED":            Cancelled,
			"DRIVER_CANCELLED":    DriverCancelled,
			"NO_DRIVER_AVAILABLE": NoDriverFound,
		},
	}
}

func (p MySam) GetOffers(ctx context.Context, _ UserInfo, s mOffer.Search, now time.Time) ([]mOffer.Offer, error) {
	reqBody := struct {
		FromLatitude          float64 `json:"fromLatitude"`
		FromLongitude         float64 `json:"fromLongitude"`
		ToLatitude            float64 `json:"toLatitude"`
		ToLongitude           float64 `json:"toLongitude"`
		NBPassengers          int     `json:"NBPassengers"`
		StartDate             string  `json:"startDate,omitempty"`
		SignificantDisability bool    `json:"significantDisability"`
	}{
		FromLatitude:          s.StartLatitude,
		FromLongitude:         s.StartLongitude,
		ToLatitude:            s.EndLatitude,
		ToLongitude:           s.EndLongitude,
		NBPassengers:          s.NbrOfPassenger,
		StartDate:             time.Now().String(),
		SignificantDisability: false,
	}

	if s.IsPlanned {
		reqBody.StartDate = s.StartDate.String()
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: [%w]", err)
	}

	req, err := p.newRequest(ctx, http.MethodPost, p.endpoint("estimation/all"), bytes.NewReader(data))

	if err != nil {
		return nil, fmt.Errorf("failed to create new request: [%w]", err)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mysam api: [%w]", err)
	}
	defer resp.Body.Close()

	//@todo think and learn about a better way to handle request responses
	switch resp.StatusCode {
	case http.StatusBadRequest:
		var data any
		json.NewDecoder(resp.Body).Decode(&data)
		return nil, fmt.Errorf("failed to fetch offer for mysam, receive following error: %v", data)
	case http.StatusOK:
		var offers []MySamOffer

		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&offers); err != nil {
			return nil, fmt.Errorf("failed to decode mysam offer format: [%w]", err)
		}

		var res []mOffer.Offer
		for _, offer := range offers {
			res = append(res, p.convertProviderOffer(offer, s, now))
		}
		return res, nil
	default:
		var data any
		json.NewDecoder(resp.Body).Decode(&data)
		return nil, fmt.Errorf("unsupported status response, receive status %v and response body %v", resp.StatusCode, data)
	}
}

func (p MySam) RequestRide(ctx context.Context, o mOffer.Offer, u UserInfo, s mOffer.Search, now time.Time) (mRide.Ride, error) {
	orderType := "IMMEDIATE"
	if o.IsPlanned {
		orderType = "RESERVATION"
	}

	reqBody := struct {
		ClientId         string       `json:"clientId"`
		FromAddress      MySamAddress `json:"fromAddress"`
		ToAddress        MySamAddress `json:"toAddress"`
		NbPassengers     int          `json:"nbPassengers"`
		PaymentMethod    string       `json:"paymentMethod"`
		WillBePaidInCash bool         `json:"willBePaidInCash"`
		VehicleType      string       `json:"vehicleType"`
		Type             string       `json:"type"`
		StartDate        string       `json:"startDate"`
	}{
		ClientId: u.MySamID,
		FromAddress: MySamAddress{
			Address:   s.StartAddress,
			Country:   s.StartCountry,
			Latitude:  s.StartLatitude,
			Longitude: s.StartLongitude,
		},
		ToAddress: MySamAddress{
			Address:   s.EndAddress,
			Country:   s.EndCountry,
			Latitude:  s.EndLatitude,
			Longitude: s.EndLongitude,
		},
		NbPassengers:     s.NbrOfPassenger,
		PaymentMethod:    "DEFERRED",
		WillBePaidInCash: false,
		VehicleType:      o.ProviderOfferName,
		Type:             orderType,
		StartDate:        o.StartDate,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return mRide.Ride{}, fmt.Errorf("failed to marshal request body: [%w]", err)
	}

	req, err := p.newRequest(ctx, http.MethodPost, p.endpoint("trips/new"), bytes.NewReader(data))
	if err != nil {
		return mRide.Ride{}, fmt.Errorf("failed to create request: [%w]", err)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return mRide.Ride{}, fmt.Errorf("failed to request ride: [%w]", err)
	}
	defer resp.Body.Close()

	var ride MySamRide
	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(&ride); err != nil {
		return mRide.Ride{}, fmt.Errorf("failed to decode response body: [%w]", err)
	}

	return p.convertProviderRide(ride, u, o, now), nil
}

func (p MySam) GetRide(ctx context.Context, ride mRide.Ride) (mRide.Info, error) {
	url := p.endpoint("trips/" + ride.ProviderRideID)

	req, err := p.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return mRide.Info{}, fmt.Errorf("failed to create request: [%w]", err)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return mRide.Info{}, fmt.Errorf("failed to fetch ride: [%w]", err)
	}
	defer resp.Body.Close()

	var updatedRide MySamRide
	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(&ride); err != nil {
		return mRide.Info{}, fmt.Errorf("failed to unmarshal ride: [%w]", err)
	}

	var driver mRide.Driver

	if updatedRide.Driver != nil {
		driver = mRide.Driver{
			DriverName:  fmt.Sprintf("%v %v", updatedRide.Driver.FirstName, updatedRide.Driver.LastName),
			DriverPhone: updatedRide.Driver.MobilePhoneNumber,
			CarPhoto:    "",
			CarLicense:  "",
		}

		if updatedRide.Driver.Location != nil {
			driver.DriverLatitude = updatedRide.Driver.Location.Latitude
			driver.DriverLongitude = updatedRide.Driver.Location.Longitude
		}

		if updatedRide.Driver.DriverDetails != nil {
			driver.CarModel = updatedRide.Driver.DriverDetails.VehicleModel
		}

	}

	return mRide.Info{
		Status:             p.StatusMapping[updatedRide.Status],
		ProviderRideID:     ride.ProviderRideID,
		ProviderStatusName: updatedRide.Status,
		Price:              updatedRide.EstimatedPrice,
		ETA:                ride.ETA,
		Driver:             driver,
	}, nil
}

func (p MySam) CancelRide(ctx context.Context, ride mRide.Ride) (mRide.Info, error) {
	url := p.endpoint(fmt.Sprintf("trips/%v/cancel", ride.ProviderRideID))

	req, err := p.newRequest(ctx, http.MethodPut, url, nil)
	if err != nil {
		return mRide.Info{}, fmt.Errorf("failed to create request: [%w]", err)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return mRide.Info{}, fmt.Errorf("failed to update ride: [%w]", err)
	}

	var updatedRide MySamRide
	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(&updatedRide); err != nil {
		return mRide.Info{}, fmt.Errorf("failed to unmarshal mysam respon [%w]", err)
	}

	return mRide.Info{
		Status:             p.StatusMapping[updatedRide.Status],
		ProviderRideID:     ride.ProviderRideID,
		ProviderStatusName: updatedRide.Status,
		Price:              updatedRide.EstimatedPrice,
		ETA:                ride.ETA,
		Driver:             ride.Driver,
	}, nil
}

func (p MySam) GetCancellationFees() {

}

func (p MySam) convertProviderOffer(offer MySamOffer, s mOffer.Search, now time.Time) mOffer.Offer {
	return mOffer.Offer{
		ID:                  validate.GenerateID(),
		StartDate:           p.convertMySamTime(offer.Estimation.StartDate).String(),
		Provider:            "mySam",
		ETA:                 offer.Estimation.Duration,
		ProviderOfferID:     fmt.Sprint(offer.Estimation.Id),
		LogoURL:             p.LogoURL,
		VehicleType:         p.OfferMapping[offer.Estimation.VehicleType],
		ProviderOfferName:   offer.Estimation.VehicleType,
		ProviderPrice:       offer.Estimation.Price,
		DisplayPrice:        fmt.Sprintf("%f %s", offer.Estimation.Price, "€"),
		DisplayPriceNumeric: offer.Estimation.Price,
		DisplayProviderName: "MySam",
		UserID:              s.UserID,
		Search:              s,
		Aggregator:          s.Aggregator,
		IsPlanned:           s.IsPlanned,
		Description:         "",
		CreatedAt:           now.String(),
		UpdatedAt:           now.String(),
		DeletedAt:           "",
	}
}

func (p MySam) convertProviderRide(ride MySamRide, u UserInfo, o mOffer.Offer, now time.Time) mRide.Ride {
	return mRide.Ride{
		ID:      validate.GenerateID(),
		UserID:  u.ID,
		OfferID: o.ID,

		IsPlanned:           o.IsPlanned,
		ETA:                 o.ETA,
		CancellationFees:    0,
		StartDate:           p.convertMySamTime(ride.StartDate),
		PaymentByTGS:        true,
		Aggregator:          o.Aggregator,
		Status:              p.StatusMapping[ride.Status],
		ProviderPrice:       ride.EstimatedPrice,
		DisplayPrice:        fmt.Sprintf("%f %s", ride.EstimatedPrice, "€"),
		DisplayPriceNumeric: ride.EstimatedPrice,
		PriceStatus:         "pending",

		Review:  mRide.Review{},
		Invoice: mRide.Invoice{},
		Payment: mRide.Payment{},
		Driver:  mRide.Driver{},

		CreatedAt: now.String(),
		UpdatedAt: now.String(),
		DeletedAt: "",
	}
}

func (p MySam) convertMySamTime(t int64) time.Time {
	return time.Unix(t/1000, (t%1000)*1000000)
}

func (p MySam) endpoint(path string) string {
	return fmt.Sprintf("%v/%v", p.BaseURL, path)
}

func (p MySam) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	fmt.Println(p.APIKey)
	req.Header.Set("X-Api-Key", p.APIKey)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
