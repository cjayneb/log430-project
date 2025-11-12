package client_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
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
	return &MatchineEngineImpl{
		BaseUrl: baseUrl + MATCHING_SERVICE_BASE_PATH,
		Client:  common.SharedHttpClient,
	}
}

func (m *MatchineEngineImpl) SubmitOrder(ctx context.Context, order *models.Order) error {
	log := common.FromContext(ctx)

	jsonData, err := json.Marshal(order)
	if err != nil {
		msg := "error marshaling JSON"
		log.Error(msg, "error", err)
		return errors.New(msg)
	}
	bodyReader := bytes.NewBuffer(jsonData)

	req, err := common.MakeAuthenticatedRequest(ctx, "POST", m.BaseUrl, bodyReader)
	if err != nil {
		msg := "error creating submit order request to matching service"
		log.Error(msg, "error", err)
		return errors.New(msg)
	}

	resp, err := m.Client.Do(req)
	if err != nil {
		msg := "error making submit order request to matching service"
		log.Error(msg, "error", err)
		return errors.New(msg)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		msg := "matching service returned non-200 status when submitting order"
		log.Error(msg, "status", resp.Status)
		return errors.New(msg)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := "error reading submit order response body"
		log.Error(msg, "error", err)
		return errors.New(msg)
	}

	var submitResp OrderSubmittedResponse
	err = json.Unmarshal(body, &submitResp)
	if err != nil {
		msg := "error unmarshaling JSON"
		log.Error(msg, "error", err)
		return errors.New(msg)
	}

	log.Info("order was successfully processed by the matching service", "response", submitResp)
	return nil
}

var _ ports.MatchingEngine = (*MatchineEngineImpl)(nil) // Ensure interface is implemented at compile time
