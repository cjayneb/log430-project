package core

import (
	"brokerx/models"
	"brokerx/ports"
	"errors"
	"fmt"
	"math"
)

const MAX_ORDER_QUANTITY int = 100000
const MIN_ORDER_QUANTITY int = 1

const INACTIVE_INSTRUMENT_MSG string = "The instrument is not active"
const INVALID_TICK_SIZE_MSG string = "The specified price does not match the instrument's tick size"
const INVALID_PRICE_BAND_MSG string = "The specified price is outside the instrument's price band"

type ComplianceService struct {
	WalletRepo ports.WalletRepository
	PositionRepo ports.PositionRepository
	MarketDataProvider ports.MarketDataProvider
}

func (service *ComplianceService) VerifyOrderCompliance(order *models.Order) error {
	if order.Quantity < MIN_ORDER_QUANTITY {
		return fmt.Errorf("order quantity must be at least %v", MIN_ORDER_QUANTITY)
	}

	if order.Quantity > MAX_ORDER_QUANTITY {
		return fmt.Errorf("order quantity surpasses maximum quantity allowed by BrokerX (%v)", MAX_ORDER_QUANTITY)
	}

	currentPrice, err := service.MarketDataProvider.GetCurrentStockPriceBySymbol(order.Symbol)
	if err != nil {
		return fmt.Errorf("error when fetching stock price {%v}: %v", order.Symbol, err)
	}
	
	if err := service.verifyBuyingPower(order, currentPrice); err != nil {
		return err
	}

	if err := service.verifySellingPower(order); err != nil {
		return err
	}

	instrument, err := service.verifyInstrumentValidity(order)
	if err != nil {
		return err
	}

	err = service.verifyTickSizeAndPriceBand(order, instrument, currentPrice)
	if err != nil {
		return err
	}

	order.Status = "open"
	return nil
}

func (service *ComplianceService) verifyBuyingPower(order *models.Order, currentPrice float64) error {
	if order.Action == "sell" {
		return nil
	}

	wallet, err := service.WalletRepo.FindByUserId(order.UserID)
	if err != nil {
		return err
	}

	if wallet.AvailableFunds < (currentPrice * float64(order.Quantity)) {
		return errors.New("not enough available funds")
	}

	return nil
}

func (service *ComplianceService) verifySellingPower(order *models.Order) error {
	if order.Action == "buy" {
		return nil
	}

	positions, err := service.PositionRepo.FindByUserIdAndSymbol(order.UserID, order.Symbol)
	if err != nil {
		return err
	}

	totalOwnedStock := 0
	for _, p := range positions {
		totalOwnedStock += p.Quantity
	}
	if totalOwnedStock < order.Quantity {
		return errors.New("not enough owned stocks")
	}

	return nil
}

func (service *ComplianceService) verifyInstrumentValidity(order *models.Order) (*models.Instrument, error) {
	instrument, err := service.MarketDataProvider.GetInstrumentBySymbol(order.Symbol)
	if err != nil {
		return nil, fmt.Errorf("error when fetching instrument {%v}: %v", order.Symbol, err)
	}

	if instrument.Status != "Active" {
		return nil, fmt.Errorf("error when verifying instrument {%v}: %v", order.Symbol, INACTIVE_INSTRUMENT_MSG)
	}

	return instrument, nil
}

func (service *ComplianceService) verifyTickSizeAndPriceBand(order *models.Order, instrument *models.Instrument, price float64) error {
	if order.Type == "market" {
		return nil
	}

	if !isValidTick(order.UnitPrice, instrument.TickSize) {
		return fmt.Errorf("error when validating tick size {%v}: %v", order.Symbol, INVALID_TICK_SIZE_MSG)
	}

	maxPrice := price + price * (1.0/float64(instrument.PriceBandPercent))
	minPrice := price - price * (1.0/float64(instrument.PriceBandPercent))
	if order.UnitPrice < minPrice || order.UnitPrice > maxPrice {
		return fmt.Errorf("error when validating price band {%v}: %v", order.Symbol, INVALID_PRICE_BAND_MSG)
	}
	
	return nil
}

func isValidTick(price, tickSize float64) bool {
    remainder := math.Mod(price, tickSize)
    return remainder < 1e-9 || math.Abs(remainder-tickSize) < 1e-9
}

var _ ports.ComplianceService = (*ComplianceService)(nil) // Ensure interface is implemented at compile time