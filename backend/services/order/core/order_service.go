package core

import (
	dao_adapters "brokerx/order-service/adapters/dao"
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)


type OrderService interface {
	PlaceOrder(ctx context.Context, order *models.Order) error
	GetOrdersForUser(ctx context.Context, userId int) ([]*models.Order, error)
}

type OrderServiceImpl struct {
	Repo              ports.OrderRepository
	OrderBook         ports.OrderBook
	EventProducer	  ports.EventProducer
	MatchingEngine    ports.MatchingEngine
}

func (service *OrderServiceImpl) PlaceOrder(ctx context.Context, order *models.Order) error {
	log := common.FromContext(ctx)

	createdOrderId, err := service.Repo.Create(ctx, order)
	if err != nil {
		log.Error("error creating order", "error", err)
		return err
	}
	order.ID = createdOrderId

	err = service.EventProducer.SendEvent(ctx, "OrderEvents", order)
	if err != nil {
		log.Error("error when sending OrderCreatedEvent", "error", err)
		return err
	}

	// if err := service.MatchingEngine.SubmitOrder(ctx, order); err != nil {
	// 	log.Error("matching failed for order", "orderId", order.ID, "error", err)
	// }
	// log.Info("Order sent to matching service", "orderId", order.ID)

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

func StartDirtyOrderSync(ctx context.Context, interval time.Duration, batchSize int, orderBook ports.OrderBook, tm ports.TransactionManager) {
	slog.Info(fmt.Sprintf("Dirty Orders interval %v seconds", interval))
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// Step 1: pop dirty order IDs from Redis
				dirtyIDs, err := popDirtyIDs(ctx, orderBook, batchSize)
				if err != nil {
					slog.Error("order sync: failed to pop dirty IDs", "error", err)
					continue
				}
				if len(dirtyIDs) == 0 {
					continue
				}

				// Step 2: fetch order data for those IDs
				orders, err := orderBook.FetchByIDs(ctx, dirtyIDs)
				if err != nil {
					slog.Error("order sync: failed to fetch orders", "error", err)
					continue
				}

				if len(orders) == 0 {
					continue
				}

				// Step 3: bulk update in MySQL
				err = tm.Do(context.Background(), func(ordersRepo ports.OrderRepository, _ ports.ExecutionRepository) error {
					return ordersRepo.UpdateBatch(ctx, orders)
				})
				if err != nil {
					slog.Error("order sync: failed to update MySQL", "error", err)
				} else {
					slog.Info(fmt.Sprintf("order sync: flushed %d dirty orders", len(orders)))
				}

			case <-ctx.Done():
				slog.Info("dirty order sync stopped")
				return
			}
		}
	}()
}

func popDirtyIDs(ctx context.Context, orderBook ports.OrderBook, batchSize int) ([]string, error) {
	rdb := orderBook.(*dao_adapters.RedisOrderBook).Rdb
	popped, err := rdb.RPopCount(ctx, "orders:dirty", batchSize).Result()
	if err == redis.Nil {
		return []string{}, nil
	}
	if err != nil {
		slog.Error("error popping dirty order ids", "error", err)
		return []string{}, err
	}
	return popped, nil
}

func PersistOrdersAndExecutions(
	ctx context.Context, 
	interval time.Duration, 
	orderBatchSize int, 
	execBatchSize int, 
	orderBook ports.OrderBook, 
	execQueue ports.ExecutionQueue, 
	tm ports.TransactionManager,
) {
	slog.Info(fmt.Sprintf("Persistence interval %v milliseconds", interval))
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ordersToPersist, err := orderBook.DequeueOrders(ctx, 100)
				if err != nil {
					slog.Error("error dequeuing orders", "error", err)
				}

				executionsToPersist, err := execQueue.DequeueExecutionRecords(ctx, 200)
				if err != nil {
					slog.Error("error dequeuing execution records", "error", err)
				}

				_ = tm.Do(ctx, func(orders ports.OrderRepository, executions ports.ExecutionRepository) error {
					if err := orders.UpdateBatch(ctx, ordersToPersist); err != nil {
						slog.Error("error saving orders", "error", err)
						return err
					}
					if err := executions.CreateBatch(executionsToPersist); err != nil {
						slog.Error("error saving execution records", "error", err)
						return err
					}
					return nil
				})

			case <-ctx.Done():
				slog.Info("Order and execution records persistance stopped")
				return
			}
		}
	}()
}

var _ OrderService = (*OrderServiceImpl)(nil) // Ensure interface is implemented at compile time
