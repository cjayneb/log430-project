package core

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
	"database/sql"
	"time"
)

type OrderService interface {
	PlaceOrder(ctx context.Context, order *models.Order) error
	GetOrdersForUser(ctx context.Context, userId int) ([]*models.Order, error)
	UpdateOrder(ctx context.Context, updatedOrder *models.Order) error
}

type OrderServiceImpl struct {
	Repo           ports.OrderRepository
	TransactionManager ports.TransactionManager
	OrderBook      ports.OrderBook
	EventProducer  ports.EventProducer
}

func (service *OrderServiceImpl) UpdateOrder(ctx context.Context, updatedOrder *models.Order) error {
	return  service.TransactionManager.Do(ctx, func(ordersRepo ports.OrderRepository, _ ports.OutboxRepository) error {
		return ordersRepo.UpdateBatch(ctx, []*models.Order{updatedOrder})
	})
}

func (service *OrderServiceImpl) PlaceOrder(ctx context.Context, order *models.Order) error {
	log := common.FromContext(ctx)

	err := service.TransactionManager.Do(ctx, func(or ports.OrderRepository, obr ports.OutboxRepository) error {
		order.Status = "open"
		createdOrderId, err := service.Repo.Create(ctx, order)
		if err != nil {
			log.Error("error creating order", "error", err)
			return obr.CreateOrderEvents(ctx, []*models.OrderEvent{{
				Topic: "OrderEvents",
				Event: "OrderCreatedFailed",
				TraceID: ctx.Value(common.CtxKeyTraceId).(string),
				Order: *order,
				Error: err.Error(),
			}})
		}
		order.ID = createdOrderId
		order.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

		return obr.CreateOrderEvents(ctx, []*models.OrderEvent{{
			Topic: "OrderEvents",
			Event: "OrderCreated",
			TraceID: ctx.Value(common.CtxKeyTraceId).(string),
			UserId: ctx.Value(common.CtxKeyUserId).(string),
			JWT: ctx.Value(common.CtxKeyJWT).(string),
			Order: *order,
		}})
	})

	if err != nil {
		log.Error("error creating order", "error", err)
		return err
	}
	return nil
}

func (service *OrderServiceImpl) GetOrdersForUser(ctx context.Context, userId int) ([]*models.Order, error) {
	log := common.FromContext(ctx)

	orders, err := service.Repo.FindByUserId(ctx, userId)
	if err != nil {
		log.Error("Error when fetching user orders", "error", err)
		return nil, err
	}

	return orders, nil
}

var _ OrderService = (*OrderServiceImpl)(nil) // Ensure interface is implemented at compile time
