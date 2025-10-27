package client_adapters

import (
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const MATCHING_SERVICE_BASE_PATH = "/api/matching/"

type MatchineEngineImpl struct {
	BaseUrl string
	Client  *http.Client
}

type SubmitOrderRequest struct {
	OrderId int `json:"order_id" validate:"required"`
}

type OrderSubmittedResponse struct {
	Message string `json:"message"`
	OrderId int    `json:"order_id"`
}

func NewMatchineEngine(baseUrl string) *MatchineEngineImpl {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	return &MatchineEngineImpl{
		BaseUrl: baseUrl + MATCHING_SERVICE_BASE_PATH,
		Client:  client,
	}
}

func (m *MatchineEngineImpl) SubmitOrder(ctx context.Context, order *models.Order) error {
	jsonData, err := json.Marshal(order)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewBuffer(jsonData)

	req, err := makeAuthenticatedRequest(ctx, "POST", m.BaseUrl, bodyReader)
	if err != nil {
		return err
	}

	resp, err := m.Client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request to matching service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("matching service returned non-200 status: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	var priceResp OrderSubmittedResponse
	err = json.Unmarshal(body, &priceResp)
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	log.Infof("order was successfully processed by the matching service : %v", priceResp)

	return nil
}

var _ ports.MatchingEngine = (*MatchineEngineImpl)(nil) // Ensure interface is implemented at compile time
