package ports

import (
	"brokerx/order-service/models"
	"context"
)

type MatchingEngine interface {
	SubmitOrder(ctx context.Context, order *models.Order) error
}
