package stripe

import (
	"fmt"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/client"
	model "vtc/business/v1/data/models"
)

type Customer struct {
	Email       string
	PhoneNumber string
	Aggregator  string
	Name        string
}

type PaymentIntent struct {
	IsThreeDSNeeded bool
	ThreeDSURL      string
	IntentID        string
	CardType        string
	PaymentMethodID string
}

type Charge struct {
	Challenge bool
	Status    stripe.PaymentIntentStatus
	URL       string
	ID        string
}

// CreateCustomer register a new stripe customer
func CreateCustomer(key string, cu Customer) (string, error) {
	sc := client.New(key, nil)

	params := &stripe.CustomerParams{
		Description: stripe.String(cu.Aggregator),
		Email:       stripe.String(cu.Email),
		Phone:       stripe.String(cu.PhoneNumber),
		Name:        stripe.String(cu.Name),
	}

	customer, err := sc.Customers.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create new stripe customer: %v", err)
	}

	return customer.ID, nil
}

// RegisterCard register a new user credit card to be used later.
func RegisterCard(key string, userStripeID string, data model.NewPaymentMethodDTO) (PaymentIntent, error) {
	sc := client.New(key, nil)

	pm, err := sc.PaymentMethods.New(&stripe.PaymentMethodParams{
		Type: stripe.String("card"),
		Card: &stripe.PaymentMethodCardParams{
			CVC:      stripe.String(data.CardCVX),
			ExpMonth: stripe.Int64(data.CardExpirationMonth),
			ExpYear:  stripe.Int64(data.CardExpirationYear),
			Number:   stripe.String(data.CardNumber),
		},
	})

	if err != nil {
		return PaymentIntent{}, fmt.Errorf("failed to register a new payment method: [%w]", err)
	}

	intent, err := sc.SetupIntents.New(&stripe.SetupIntentParams{
		Customer:           stripe.String(userStripeID),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
	})
	if err != nil {
		return PaymentIntent{}, fmt.Errorf("failed to setup a intent: [%w]", err)
	}

	resp, err := sc.SetupIntents.Confirm(
		intent.ID,
		&stripe.SetupIntentConfirmParams{PaymentMethod: stripe.String(pm.ID), ReturnURL: stripe.String(data.ReturnUrl)},
	)
	if err != nil {
		return PaymentIntent{}, fmt.Errorf("failed to confirm the intent: [%w]", err)
	}

	var isThreeDSNeeded bool
	var ThreeDSURL string

	if resp.NextAction != nil {
		if resp.NextAction.RedirectToURL != nil {
			isThreeDSNeeded = len(resp.NextAction.RedirectToURL.URL) > 0
			ThreeDSURL = resp.NextAction.RedirectToURL.URL
		}
	}

	return PaymentIntent{
		IsThreeDSNeeded: isThreeDSNeeded,
		ThreeDSURL:      ThreeDSURL,
		IntentID:        intent.ID,
		PaymentMethodID: pm.ID,
		CardType:        string(pm.Card.Brand),
	}, nil
}

// CancelPayment cancel the given payment
func CancelPayment(key, preAuthID, reason string) bool {
	sc := client.New(key, nil)

	resp, err := sc.PaymentIntents.Cancel(preAuthID, &stripe.PaymentIntentCancelParams{CancellationReason: stripe.String(reason)})
	if err != nil {
		return false
	}

	return resp.Status == stripe.PaymentIntentStatusCanceled
}

// CreateCharge create a new payment, the capture method is manual, so you will need to call CapturePayment to finalize the process
func CreateCharge(key string, amount float64, userStripeID, paymentMethodID, returnURL, currency string) (Charge, error) {
	sc := client.New(key, nil)

	intent, err := sc.PaymentIntents.New(&stripe.PaymentIntentParams{
		Amount:        stripe.Int64(int64(amount) * 100),
		Customer:      stripe.String(userStripeID),
		PaymentMethod: stripe.String(paymentMethodID),
		OffSession:    stripe.Bool(true),
		Confirm:       stripe.Bool(true),
		ReturnURL:     stripe.String(returnURL),
		CaptureMethod: stripe.String("manual"),
		Currency:      stripe.String(currency),
	})

	if err != nil {
		return Charge{}, fmt.Errorf("failed to create a new payment intent: [%w]", err)
	}

	var url string
	var challenge bool

	if intent.NextAction != nil {
		if intent.NextAction.RedirectToURL != nil {
			url = intent.NextAction.RedirectToURL.URL
			challenge = len(intent.NextAction.RedirectToURL.URL) > 0
		}
	}

	return Charge{
		ID:        intent.ID,
		URL:       url,
		Status:    intent.Status,
		Challenge: challenge,
	}, nil
}

// CapturePayment capture the given amount for the payment. If the amount is inferior to the blocked amount, the remaining
// sum will be refund
func CapturePayment(key, preAuthID string, amount int64) error {
	sc := client.New(key, nil)

	pi, err := sc.PaymentIntents.Get(preAuthID, nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve givent payment: [%w]", err)
	}

	if pi.Status == stripe.PaymentIntentStatusCanceled || pi.Status == stripe.PaymentIntentStatusSucceeded {
		return nil
	}

	if _, err := sc.PaymentIntents.Capture(pi.ID, &stripe.PaymentIntentCaptureParams{
		AmountToCapture: stripe.Int64(amount * 100),
	}); err != nil {
		return fmt.Errorf("failed to capture payment: [%w]", err)
	}

	return nil
}

// GetPaymentStatus retrieve the status of the given payment
func GetPaymentStatus(key, id string) (stripe.PaymentIntentStatus, error) {
	sc := client.New(key, nil)

	pi, err := sc.PaymentIntents.Get(id, nil)
	if err != nil {
		return stripe.PaymentIntentStatusCanceled, fmt.Errorf("failed to retrieve the given payment: [%w]", err)
	}

	return pi.Status, nil
}
