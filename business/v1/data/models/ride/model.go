package ride

import "time"

// Ride represent a ride order by a user
type Ride struct {
	ID             string `json:"id" bson:"_id"`
	UserID         string `json:"userID" bson:"userID"`
	OfferID        string `json:"offerID" bson:"offerID"`
	ProviderRideID string `json:"providerRideID" bson:"providerRideID"`

	IsPlanned        bool      `json:"isPlanned" bson:"isPlanned"`
	ETA              float64   `json:"ETA" bson:"ETA"`
	CancellationFees float64   `json:"cancellationFees" bson:"cancellationFees"`
	StartDate        time.Time `json:"startDate" bson:"startDate"`
	PaymentByTGS     bool      `json:"paymentByTGS" bson:"paymentByTGS"`
	Aggregator       string    `json:"aggregator" bson:"aggregator"`
	Status           string    `json:"status" bson:"status"`

	ProviderPrice       float64 `json:"providerPrice" bson:"providerPrice"`
	DisplayPrice        string  `json:"displayPrice" bson:"displayPrice"`
	DisplayPriceNumeric float64 `json:"displayPriceNumeric" bson:"displayPriceNumeric"`
	PriceStatus         string  `json:"priceStatus" bson:"priceStatus"`

	Review  Review  `json:"review" bson:"review"`
	Invoice Invoice `json:"invoice" bson:"invoice"`
	Payment Payment `json:"payment" bson:"payment"`
	Driver  Driver  `json:"driver" bson:"driver"`

	CreatedAt string `json:"createdAt" bson:"createdAt"`
	UpdatedAt string `json:"updatedAt" bson:"updatedAt"`
	DeletedAt string `json:"deletedAt" bson:"deletedAt"`
}

// Driver represent a driver assign to a ride
type Driver struct {
	DriverName      string  `json:"driverName" bson:"driverName"`
	DriverPhone     string  `json:"driverPhone" bson:"driverPhone"`
	DriverLatitude  float64 `json:"driverLatitude" bson:"driverLatitude"`
	DriverLongitude float64 `json:"driverLongitude" bson:"driverLongitude"`

	CarModel   string `json:"carModel" bson:"carModel"`
	CarPhoto   string `json:"carPhoto" bson:"carPhoto"`
	CarLicense string `json:"carLicense" bson:"carLicense"`
}

// Review represent a review made by a user
type Review struct {
	Rating float64 `json:"rating"`

	CreatedAt string `json:"createdAt" bson:"createdAt"`
	UpdatedAt string `json:"updatedAt" bson:"updatedAt"`
	DeletedAt string `json:"deletedAt" bson:"deletedAt"`
}

// Invoice represent a invoice generate for a ride
type Invoice struct {
	Amount          float64   `json:"amount" bson:"amount"`
	InvoiceFileOnS3 string    `json:"invoiceFileOnS3" bson:"invoiceFileOnS3"`
	Date            time.Time `json:"date" bson:"date"`
	Nature          string    `json:"nature" bson:"nature"`
	To              string    `json:"to" bson:"to"`
	From            string    `json:"from" bson:"from"`
	AddressTo       string    `json:"addressTo" bson:"addressTo"`

	CreatedAt string `json:"createdAt" bson:"createdAt"`
	UpdatedAt string `json:"updatedAt" bson:"updatedAt"`
	DeletedAt string `json:"deletedAt" bson:"deletedAt"`
}

// Payment represent a payment made by a user to pay a ride
type Payment struct {
	Date            time.Time `json:"date" bson:"date"`
	Status          string    `json:"status" bson:"status"`
	PreAuthID       string    `json:"preAuthID" bson:"preAuthID"`
	ThreeDsURL      string    `json:"threeDsURL" bson:"threeDsURL"`
	PreAuthPrice    float64   `json:"preAuthPrice" bson:"preAuthPrice"`
	Challenge       bool      `json:"challenge" bson:"challenge"`
	PaymentMethodID string    `json:"paymentMethodID" bson:"paymentMethodID"`

	CreatedAt string `json:"createdAt" bson:"createdAt"`
	UpdatedAt string `json:"updatedAt" bson:"updatedAt"`
	DeletedAt string `json:"deletedAt" bson:"deletedAt"`
}

// Info represent the data that change during the execution of the ride
type Info struct {
	Status             string
	ProviderRideID     string
	ProviderStatusName string
	Price              float64
	ETA                float64
	Driver             Driver
}

// CreatePaymentDTO create a new payment for a ride
type CreatePaymentDTO struct {
	ReturnURL      string `json:"returnURL" validate:"required"`
	OfferID        string `json:"offerID" validate:"required"`
	UserID         string `json:"userID" validate:"required"`
	AggregatorCode string `json:"aggregatorCode" validate:"required"`
}
