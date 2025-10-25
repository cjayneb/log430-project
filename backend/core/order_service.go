package core

import (
	"brokerx/adapters"
	"brokerx/models"
	"brokerx/ports"
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

var orderQueueLen = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "order_queue_length",
	Help: "Current number of orders in the matching queue",
})

func init() { prometheus.MustRegister(orderQueueLen) }

var orderQueue chan *models.Order

func (service *OrderService) StartMatchingWorkers() {
	orderQueue = make(chan *models.Order, 1000)
	for i := 0; i < 8; i++ {
		go func() {
			for order := range orderQueue {
				if err := service.OrderBook.Insert(order); err != nil {
					// TODO: retry? or find a way to let user know
					log.Errorf("Order book submission failed for order #%d: %v", order.ID, err)
					order.Status = "canceled"
					if err = service.Repo.Update(order); err != nil {
                        log.Errorf("Failed to update canceled order #%d: %v", order.ID, err)
                    }
					return
				}
				if err := service.MatchingEngine.SubmitOrder(order.ID); err != nil {
					log.Errorf("matching failed for order #%d: %v", order.ID, err)
				}
			}
		}()
	}
	log.Infof("Started %d matching workers", 8)
}

func StartDirtyOrderSync(ctx context.Context, interval time.Duration, batchSize int, orderBook ports.OrderBook, tm ports.TransactionManager) {
    ticker := time.NewTicker(interval)
    go func() {
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                // Step 1: pop dirty order IDs from Redis
                dirtyIDs, err := popDirtyIDs(ctx, orderBook, batchSize)
                if err != nil {
                    log.Errorf("order sync: failed to pop dirty IDs: %v", err)
                    continue
                }
                if len(dirtyIDs) == 0 {
                    continue
                }

                // Step 2: fetch order data for those IDs
                orders, err := orderBook.FetchByIDs(dirtyIDs)
                if err != nil {
                    log.Errorf("order sync: failed to fetch orders: %v", err)
                    continue
                }

                if len(orders) == 0 {
                    continue
                }

                // Step 3: bulk update in MySQL
                err = tm.Do(context.Background(), func(ordersRepo ports.OrderRepository, _ ports.ExecutionRepository) error {
                    return ordersRepo.UpdateBatch(orders)
                })
                if err != nil {
                    log.Errorf("order sync: failed to update MySQL: %v", err)
                } else {
                    log.Infof("order sync: flushed %d dirty orders", len(orders))
                }

            case <-ctx.Done():
                log.Info("dirty order sync stopped")
                return
            }
        }
    }()
}

// pop up to batchSize dirty IDs
func popDirtyIDs(ctx context.Context, orderBook ports.OrderBook, batchSize int) ([]string, error) {
    rdb := orderBook.(*adapters.RedisOrderBook).Rdb
    pipe := rdb.Pipeline()
    cmds := make([]*redis.StringCmd, batchSize)
    for i := 0; i < batchSize; i++ {
        cmds[i] = pipe.RPop(ctx, "orders:dirty")
    }
    _, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        return nil, err
    }

    ids := []string{}
    for _, cmd := range cmds {
        id, err := cmd.Result()
        if err == nil && id != "" {
            ids = append(ids, id)
        }
    }
    return ids, nil
}


type OrderService struct {
	Repo              ports.OrderRepository
	ComplianceService ports.ComplianceService
	OrderBook 		  ports.OrderBook
	MatchingEngine    ports.MatchingEngine
}

func (service *OrderService) PlaceOrder(order *models.Order) error {
	err := service.ComplianceService.VerifyOrderCompliance(order)
	if err != nil {
		log.Errorf("Error when verifying order compliance : %v", err)
		return err
	}

	createdOrderId, err := service.Repo.Create(order)
	if err != nil {
		return err
	}
	order.ID = createdOrderId

	orderQueueLen.Set(float64(len(orderQueue)))
	orderQueue <- order
	log.Infof("Order #%v queued for matching", order.ID)
    
	return nil
}

func (service *OrderService) GetOrdersForUser(userId string) ([]*models.Order, error) {
	orders, err := service.Repo.FindByUserId(userId)
	if err != nil {
		log.Errorf("Error when fetching user orders : %v", err)
		return nil, err
	}

	return orders, nil
}

var _ ports.OrderService = (*OrderService)(nil) // Ensure interface is implemented at compile time
