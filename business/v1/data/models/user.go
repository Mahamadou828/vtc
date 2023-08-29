package models

// User represent an individual user
type User struct {
	ID               string          `bson:"_id" json:"id"`
	Email            string          `bson:"email" json:"email"`
	PhoneNumber      string          `bson:"phoneNumber" json:"phoneNumber"`
	Name             string          `bson:"name" json:"name"`
	StripeID         string          `bson:"stripeID" json:"stripeID"`
	MySamClientID    string          `bson:"mySamClientID" json:"mySamClientID"`
	Aggregator       string          `bson:"aggregator" json:"aggregator"`
	PushSubscription string          `bson:"pushSubscription" json:"pushSubscription"`
	CognitoID        string          `bson:"cognitoID" json:"cognitoID"`
	Addresses        []Address       `bson:"addresses" json:"addresses"`
	PaymentMethods   []PaymentMethod `bson:"paymentMethods" json:"paymentMethods"`
	CreatedAt        string          `bson:"createdAt" json:"createdAt"`
	UpdatedAt        string          `bson:"updatedAt" json:"updatedAt"`
	DeletedAt        string          `bson:"deletedAt" json:"deletedAt"`
}

// Address represent a favorite address save by the user
type Address struct {
	Address   string  `bson:"address" json:"address"`
	Country   string  `bson:"country" json:"country"`
	Latitude  float64 `bson:"latitude" json:"latitude"`
	Longitude float64 `bson:"longitude" json:"longitude"`
	Type      string  `bson:"type" json:"type"`
	CreatedAt string  `bson:"createdAt" json:"createdAt"`
	UpdatedAt string  `bson:"updatedAt" json:"updatedAt"`
	DeletedAt string  `bson:"deletedAt" json:"deletedAt"`
}

// PaymentMethod represent a credit card use by a user to pay drive
type PaymentMethod struct {
	Name              string `bson:"name" json:"name"`
	Active            bool   `bson:"active" json:"active"`
	CreditCardPayload string `bson:"creditCardPayload" json:"creditCardPayload"`
	IntentID          string `bson:"intentID" json:"intentID"`
	PaymentServiceID  string `bson:"paymentServiceID" json:"paymentServiceID"`
	CreditCardType    string `bson:"creditCardType" json:"creditCardType"`
	IsFavorite        bool   `bson:"isFavorite" json:"isFavorite"`
	CreatedAt         string `bson:"createdAt" json:"createdAt"`
	UpdatedAt         string `bson:"updatedAt" json:"updatedAt"`
	DeletedAt         string `bson:"deletedAt" json:"deletedAt"`
}

// NewUserDTO define all information needed to create a new user account
type NewUserDTO struct {
	Email       string `json:"email,omitempty" validate:"email,required"`
	PhoneNumber string `json:"phoneNumber,omitempty" validate:"required"`
	Name        string `json:"name,omitempty" validate:"required"`
	Password    string `json:"password,omitempty" validate:"required"`
}

// LoginDTO define all information to log a user
type LoginDTO struct {
	Email    string `json:"email" validate:"email,required"`
	Password string `json:"password" validate:"required"`
}

// NewPaymentMethodDTO represent all data needed to create a new user payment method to pay for rides
type NewPaymentMethodDTO struct {
	CardNumber          string `json:"cardNumber" validate:"required"`
	CardExpirationYear  int64  `json:"cardExpirationYear" validate:"required"`
	CardExpirationMonth int64  `json:"cardExpirationMonth" validate:"required"`
	ReturnUrl           string `json:"returnUrl" validate:"required"`
	PaymentMethodName   string `json:"paymentMethodName" validate:"required"`
	CardCVX             string `json:"cardCVX" validate:"required"`
	UserID              string `json:"userID" validate:"required"`
	IsFavorite          bool   `json:"isFavorite" validate:"required"`
}
