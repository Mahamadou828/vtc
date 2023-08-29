package provider

import (
	"context"
	"time"

	model "vtc/business/v1/data/models"
	"vtc/foundation/config"
)

func CreatePayment(ctx context.Context, data model.CreatePaymentDTO, cfg *config.App, now time.Time) {
}

func RequestRide() {}

func GetRide() {}

func RefreshRide() {}

func CancelRide() {}
