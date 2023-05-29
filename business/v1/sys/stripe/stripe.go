package stripe

import (
	"fmt"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/client"
)

type Customer struct {
	Email       string
	PhoneNumber string
	Aggregator  string
	Name        string
}

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
