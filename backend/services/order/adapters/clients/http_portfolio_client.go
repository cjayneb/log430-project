package client_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const PORTFOLIO_SERVICE_BASE_PATH = "/api/portfolio"
const WALLET_ENDPOINT = "/wallet"
const POSITIONS_ENDPOINT = "/positions"

type PortfolioServiceImpl struct {
	BaseUrl string
	Client  *http.Client
}

type PortfolioWalletResponse struct {
	Wallet *models.Wallet `json:"wallet"`
}

type PortfolioPositionsResponse struct {
	Positions []*models.Position `json:"positions"`
}

func NewPortfolioServiceClient(baseUrl string) *PortfolioServiceImpl {
	return &PortfolioServiceImpl{
		BaseUrl: baseUrl + PORTFOLIO_SERVICE_BASE_PATH,
		Client:  common.SharedHttpClient,
	}
}

func (p *PortfolioServiceImpl) FetchPositions(ctx context.Context, userId int, symbol string) ([]*models.Position, error) {
	log := common.FromContext(ctx)

	req, err := common.MakeAuthenticatedRequest(ctx, "GET", p.BaseUrl+POSITIONS_ENDPOINT+"?symbol="+symbol, nil)
	if err != nil {
		msg := "error creating request to fetch positions from portfolio service"
		log.Error(msg, "error", err)
		return nil, errors.New(msg)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		msg := "error making positions request to portfolio service"
		log.Error(msg, "error", err)
		return nil, errors.New(msg)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg := "portfolio service returned non-200 status when fetching positions"
		log.Error(msg, "status", resp.Status)
		return nil, errors.New(msg)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := "error reading positions response body"
		log.Error(msg, "error", err)
		return nil, errors.New(msg)
	}

	var positionsResp PortfolioPositionsResponse
	err = json.Unmarshal(body, &positionsResp)
	if err != nil {
		msg := "error unmarshaling JSON"
		log.Error(msg, "error", err)
		return nil, errors.New(msg)
	}

	return positionsResp.Positions, nil
}

func (p *PortfolioServiceImpl) GetWallet(ctx context.Context, userId int) (*models.Wallet, error) {
	log := common.FromContext(ctx)

	req, err := common.MakeAuthenticatedRequest(ctx, "GET", p.BaseUrl+WALLET_ENDPOINT, nil)
	if err != nil {
		msg := "error creating request to fetch wallet from portfolio service"
		log.Error(msg, "error", err)
		return nil, errors.New(msg)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		msg := "error making wallet request to portfolio service"
		log.Error(msg, "error", err)
		return nil, errors.New(msg)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg := "portfolio service returned non-200 status when fetching wallet"
		log.Error(msg, "status", resp.Status)
		return nil, errors.New(msg)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := "error reading wallet response body"
		log.Error(msg, "error", err)
		return nil, errors.New(msg)
	}

	var walletResp PortfolioWalletResponse
	err = json.Unmarshal(body, &walletResp)
	if err != nil {
		msg := "error unmarshaling JSON"
		log.Error(msg, "error", err)
		return nil, errors.New(msg)
	}

	return walletResp.Wallet, nil
}

var _ ports.PortfolioService = (*PortfolioServiceImpl)(nil) // Ensure interface is implemented at compile time
