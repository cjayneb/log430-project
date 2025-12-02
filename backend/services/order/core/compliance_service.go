package core

import (
	"brokerx/order-service/common"
	"brokerx/order-service/core/rules"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
	"fmt"
)

type ComplianceService interface {
	VerifyOrderCompliance(ctx context.Context, order *models.Order) error
}

type ComplianceServiceImpl struct {
	MarketDataProvider ports.MarketDataProvider
	complianceRules    []rules.ComplianceRule
}

func NewComplianceService(marketDataProvider ports.MarketDataProvider) *ComplianceServiceImpl {
	complianceRules := []rules.ComplianceRule{
		rules.NewCR001OrderQuantity(),
		rules.NewCR002InstrumentValidity(),
		rules.NewCR003InstrumentTickSize(),
		rules.NewCR004InstrumentPriceBand(),
	}
	return &ComplianceServiceImpl{
		MarketDataProvider: marketDataProvider,
		complianceRules:    complianceRules,
	}
}

func (service *ComplianceServiceImpl) VerifyOrderCompliance(ctx context.Context, order *models.Order) error {
	log := common.FromContext(ctx)

	currentPrice, err := service.MarketDataProvider.GetCurrentStockPriceBySymbol(ctx, order.Symbol)
	if err != nil {
		e := fmt.Errorf("%w: error when fetching stock price {%v}: %v", common.ErrDependencyFailure, order.Symbol, err)
		log.Error(e.Error())
		return e
	}
	instrument, err := service.MarketDataProvider.GetInstrumentBySymbol(ctx, order.Symbol)
	if err != nil {
		e := fmt.Errorf("%w: error when fetching instrument {%v}: %v", common.ErrDependencyFailure, order.Symbol, err)
		log.Error(e.Error())
		return e
	}
	inputs := rules.ComplianceRuleInputs{
		Instrument:   instrument,
		CurrentPrice: &currentPrice,
	}

	for _, rule := range service.complianceRules {
		if err := rule.Setup(inputs); err != nil {
			log.Error(err.Error())
			return err
		}
		if err := rule.Verify(ctx, order); err != nil {
			log.Error(err.Error())
			return err
		}
	}

	order.Status = "open"
	return nil
}

var _ ComplianceService = (*ComplianceServiceImpl)(nil) // Ensure interface is implemented at compile time
