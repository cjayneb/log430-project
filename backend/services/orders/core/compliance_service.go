package core

import (
	"brokerx/order-service/core/rules"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"fmt"
)

type ComplianceService interface {
	VerifyOrderCompliance(order *models.Order) error
}

type ComplianceServiceImpl struct {
	PortfolioService   ports.PortfolioService
	MarketDataProvider ports.MarketDataProvider
	complianceRules    []rules.ComplianceRule
}

func NewComplianceService(portfolioService ports.PortfolioService, marketDataProvider ports.MarketDataProvider) *ComplianceServiceImpl {
	complianceRules := []rules.ComplianceRule{
		rules.NewCR001OrderQuantity(),
		rules.NewCR002BuyingPower(portfolioService),
		rules.NewCR003SellingPower(portfolioService),
		rules.NewCR004InstrumentValidity(),
		rules.NewC005InstrumentTickSize(),
		rules.NewCR006InstrumentPriceBand(),
	}
	return &ComplianceServiceImpl{
		PortfolioService:   portfolioService,
		MarketDataProvider: marketDataProvider,
		complianceRules:    complianceRules,
	}
}

func (service *ComplianceServiceImpl) VerifyOrderCompliance(order *models.Order) error {
	currentPrice, err := service.MarketDataProvider.GetCurrentStockPriceBySymbol(order.Symbol)
	if err != nil {
		return fmt.Errorf("error when fetching stock price {%v}: %v", order.Symbol, err)
	}
	instrument, err := service.MarketDataProvider.GetInstrumentBySymbol(order.Symbol)
	if err != nil {
		return fmt.Errorf("error when fetching instrument {%v}: %v", order.Symbol, err)
	}
	inputs := rules.ComplianceRuleInputs{
		Instrument:   instrument,
		CurrentPrice: currentPrice,
	}

	for _, rule := range service.complianceRules {
		if err := rule.Setup(inputs); err != nil {
			return err
		}
		if err := rule.Verify(order); err != nil {
			return err
		}
	}

	order.Status = "open"
	return nil
}

var _ ComplianceService = (*ComplianceServiceImpl)(nil) // Ensure interface is implemented at compile time
