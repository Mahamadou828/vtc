package models

import "time"

// Offer represent an offer return by a provider
type Offer struct {
	ID                  string  `bson:"_id" json:"id,omitempty"`
	StartDate           string  `json:"startDate" bson:"startDate"`
	Provider            string  `json:"provider" bson:"provider"`
	ETA                 float64 `json:"ETA" bson:"ETA"`
	ProviderOfferID     string  `json:"providerOfferID" bson:"providerOfferID"`
	LogoURL             string  `json:"logoURL" bson:"logoURL"`
	VehicleType         string  `json:"vehicleType" bson:"vehicleType"`
	ProviderOfferName   string  `json:"providerOfferName" bson:"providerOfferName"`
	ProviderPrice       float64 `json:"providerPrice" bson:"providerPrice"`
	DisplayPrice        string  `json:"displayPrice" bson:"displayPrice"`
	DisplayPriceNumeric float64 `json:"displayPriceNumeric" bson:"displayPriceNumeric"`
	DisplayProviderName string  `json:"displayProviderName" bson:"displayProviderName"`
	IsPlanned           bool    `json:"isPlanned" bson:"isPlanned"`
	UserID              string  `json:"userID" bson:"userID"`
	Search              Search  `json:"search" bson:"search"`
	Aggregator          string  `json:"aggregator" bson:"aggregator"`
	Description         string  `json:"description" bson:"description"`
	CreatedAt           string  `bson:"createdAt" json:"createdAt"`
	UpdatedAt           string  `bson:"updatedAt" json:"updatedAt"`
	DeletedAt           string  `bson:"deletedAt" json:"deletedAt"`
}

// Search represent a search make to fetch offer make by a user
type Search struct {
	ID         string `json:"id" bson:"_id"`
	UserID     string `json:"userID" bson:"userID"`
	Aggregator string `json:"aggregator" bson:"aggregator"`

	StartDate     time.Time `json:"startDate" bson:"startDate"`
	AskedProvider []string  `json:"askedProvider" bson:"askedProvider"`

	StartAddress   string  `json:"startAddress" bson:"startAddress"`
	StartLatitude  float64 `json:"startLatitude" bson:"startLatitude"`
	StartLongitude float64 `json:"startLongitude" bson:"startLongitude"`
	StartCountry   string  `json:"startCountry" bson:"startCountry"`

	EndAddress   string  `json:"endAddress" bson:"endAddress"`
	EndLatitude  float64 `json:"endLatitude" bson:"endLatitude"`
	EndLongitude float64 `json:"endLongitude" bson:"endLongitude"`
	EndCountry   string  `json:"endCountry" bson:"endCountry"`

	Distance       float64 `json:"distance" bson:"distance"`
	NbrOfPassenger int     `json:"nbrOfPassenger" bson:"nbrOfPassenger"`
	IsPlanned      bool    `json:"isPlanned" bson:"isPlanned"`

	CreatedAt string `json:"createdAt" bson:"createdAt"`
	UpdatedAt string `json:"updatedAt" bson:"updatedAt"`
	DeletedAt string `json:"deletedAt" bson:"deletedAt"`
}

// GetOfferDTO define all data needed to fetch offer from providers
type GetOfferDTO struct {
	UserID    string `json:"userID" validate:"required"`
	StartDate string `json:"startDate,omitempty"`

	StartAddress   string  `json:"startAddress" validate:"required"`
	StartLatitude  float64 `json:"startLatitude" validate:"required,latitude"`
	StartLongitude float64 `json:"startLongitude" validate:"required,longitude"`
	StartCountry   string  `json:"startCountry" validate:"required"`

	EndAddress   string  `json:"endAddress" validate:"required"`
	EndLatitude  float64 `json:"endLatitude" validate:"required,latitude"`
	EndLongitude float64 `json:"endLongitude" validate:"required,longitude"`
	EndCountry   string  `json:"endCountry" validate:"required"`

	Distance       float64  `json:"distance" validate:"required"`
	NbrOfPassenger int      `json:"nbrOfPassenger" validate:"required"`
	ProviderList   []string `json:"providerList" validate:"required"`
}
