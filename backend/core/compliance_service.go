package core

import (
	"brokerx/models"
	"brokerx/ports"
	"errors"
)

type ComplianceService struct {
	WalletRepo ports.WalletRepository
	PositionRepo ports.PositionRepository
}

func (service *ComplianceService) VerifyOrderCompliance(order *models.Order) error {

	if order.Action == "buy" {
		if err := service.verifyBuyOrderCompliance(order); err != nil {
			return err
		}
	}

	if order.Action == "sell" {
		if err := service.verifySellOrderCompliance(order); err != nil {
			return err
		}
	}

	return nil
}

func (service *ComplianceService) verifyBuyOrderCompliance(order *models.Order) error {
	wallet, err := service.WalletRepo.FindByUserId(order.UserID)
	if err != nil {
		return err
	}

	if wallet.AvailableFunds < (order.UnitPrice * float64(order.Quantity)) {
		return errors.New("not enough available funds")
	}

	return nil
}

func (service *ComplianceService) verifySellOrderCompliance(order *models.Order) error {
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

var _ ports.ComplianceService = (*ComplianceService)(nil) // Ensure interface is implemented at compile time