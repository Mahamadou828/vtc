package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
	"vtc/business/v1/data/models"

	"vtc/business/v1/sys/validate"
	"vtc/foundation/config"
)

var (
	MySAMFailedToFetchApi       = errors.New("failed to fetch mysam api")
	MySAMUnexpectedResponseBody = errors.New("failed to decode response body")
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

func (p MySam) GetOffers(ctx context.Context, _ UserInfo, s models.Search, now time.Time) ([]models.Offer, error) {
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
		return nil, fmt.Errorf("%w: %v", ErrFailedToMarshalRequest, err)
	}

	req, err := p.newRequest(ctx, http.MethodPost, p.endpoint("estimation/all"), bytes.NewReader(data))

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToCreateRequest, err)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", MySAMFailedToFetchApi, err)
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
			return nil, fmt.Errorf("%w: %v", MySAMUnexpectedResponseBody, err)
		}

		var res []models.Offer
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

func (p MySam) RequestRide(ctx context.Context, o models.Offer, u UserInfo, s models.Search, now time.Time) (models.Ride, error) {
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
		return models.Ride{}, fmt.Errorf("%w: %v", ErrFailedToMarshalRequest, err)
	}

	req, err := p.newRequest(ctx, http.MethodPost, p.endpoint("trips/new"), bytes.NewReader(data))
	if err != nil {
		return models.Ride{}, fmt.Errorf("%w: %v", ErrFailedToCreateRequest, err)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return models.Ride{}, fmt.Errorf("%w: %v", MySAMFailedToFetchApi, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var ride MySamRide
		decoder := json.NewDecoder(resp.Body)

		if err := decoder.Decode(&ride); err != nil {
			return models.Ride{}, fmt.Errorf("%w: %v", MySAMUnexpectedResponseBody, err)
		}

		return p.convertProviderRide(ride, u, o, now), nil
	case http.StatusBadRequest:
		var data any
		json.NewDecoder(resp.Body).Decode(&data)
		return models.Ride{}, fmt.Errorf("failed to fetch offer for mysam, receive following error: %v", data)
	default:
		var data any
		json.NewDecoder(resp.Body).Decode(&data)
		return models.Ride{}, fmt.Errorf("unsupported status response, receive status %v and response body %v", resp.StatusCode, data)
	}

}

func (p MySam) GetRide(ctx context.Context, ride models.Ride) (models.Info, error) {
	url := p.endpoint("trips/" + ride.ProviderRideID)

	req, err := p.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return models.Info{}, fmt.Errorf("%w: %v", ErrFailedToCreateRequest, err)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return models.Info{}, fmt.Errorf("%w: %v", MySAMFailedToFetchApi, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var data any
		json.NewDecoder(resp.Body).Decode(&data)
		return models.Info{}, fmt.Errorf("no ride found: %v", data)
	}

	var updatedRide MySamRide
	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(&ride); err != nil {
		return models.Info{}, fmt.Errorf("failed to unmarshal ride: [%w]", err)
	}

	var driver models.Driver

	if updatedRide.Driver != nil {
		driver = models.Driver{
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

	return models.Info{
		Status:             p.StatusMapping[updatedRide.Status],
		ProviderRideID:     ride.ProviderRideID,
		ProviderStatusName: updatedRide.Status,
		Price:              updatedRide.EstimatedPrice,
		ETA:                ride.ETA,
		Driver:             driver,
	}, nil
}

func (p MySam) CancelRide(ctx context.Context, ride models.Ride) (models.Info, error) {
	url := p.endpoint(fmt.Sprintf("trips/%v/cancel", ride.ProviderRideID))

	req, err := p.newRequest(ctx, http.MethodPut, url, nil)
	if err != nil {
		return models.Info{}, fmt.Errorf("%w: %v", ErrFailedToCreateRequest, err)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return models.Info{}, fmt.Errorf("failed to update ride: [%w]", err)
	}

	var updatedRide MySamRide
	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(&updatedRide); err != nil {
		return models.Info{}, fmt.Errorf("failed to unmarshal mysam response [%w]", err)
	}

	return models.Info{
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

func (p MySam) convertProviderOffer(offer MySamOffer, s models.Search, now time.Time) models.Offer {
	return models.Offer{
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

func (p MySam) convertProviderRide(ride MySamRide, u UserInfo, o models.Offer, now time.Time) models.Ride {
	return models.Ride{
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

		Review:  models.Review{},
		Invoice: models.Invoice{},
		Payment: models.Payment{},
		Driver:  models.Driver{},

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
