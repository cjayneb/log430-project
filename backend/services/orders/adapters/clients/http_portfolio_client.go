package client_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
	"encoding/json"
	"fmt"
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
	req, err := makeAuthenticatedRequest(ctx, "GET", p.BaseUrl+POSITIONS_ENDPOINT+"?symbol="+symbol, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to portfolio service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("portfolio service returned non-200 status: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var positionsResp PortfolioPositionsResponse
	err = json.Unmarshal(body, &positionsResp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return positionsResp.Positions, nil
}

func (p *PortfolioServiceImpl) GetWallet(ctx context.Context, userId int) (*models.Wallet, error) {
	req, err := makeAuthenticatedRequest(ctx, "GET", p.BaseUrl+WALLET_ENDPOINT, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to portfolio service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("portfolio service returned non-200 status: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var walletResp PortfolioWalletResponse
	err = json.Unmarshal(body, &walletResp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return walletResp.Wallet, nil
}

func makeAuthenticatedRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	jwt, ok := ctx.Value(common.CtxKeyJWT).(string)
	if !ok {
		return nil, fmt.Errorf("missing JWT in context")
	}
	userId, ok := ctx.Value(common.CtxKeyUserId).(string)
	if !ok {
		return nil, fmt.Errorf("missing JWT in context")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request : %v", err)
	}
	req.Header.Set(common.HeaderKeyAuth, jwt)
	req.Header.Set(common.HeaderKeyUserId, userId)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

var _ ports.PortfolioService = (*PortfolioServiceImpl)(nil) // Ensure interface is implemented at compile time
