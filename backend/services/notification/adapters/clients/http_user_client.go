package client_adapters

import (
	"brokerx/notification-service/models"
	"brokerx/notification-service/ports"
	"brokerx/notification-service/util"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

const USER_SERVICE_BASE_PATH = "/api/user"
const GET_USER_ENDPOINT = "/contact"

type UserServiceClient struct {
	BaseUrl string
	Client  *http.Client
}

func NewUserServiceImpl(baseUrl string) *UserServiceClient {
	return &UserServiceClient{
		BaseUrl: baseUrl + USER_SERVICE_BASE_PATH,
		Client:  util.SharedHttpClient,
	}
}

func (u *UserServiceClient) GetUserContactInfo(ctx context.Context, userID int) (models.UserContactInfo, error) {
	log := util.FromContext(ctx)

	resp, err := util.MakeAuthenticatedRequest(ctx, "GET", u.BaseUrl+GET_USER_ENDPOINT+"?userId="+strconv.Itoa(userID), nil)
	if err != nil {
		msg := "error making request"
		log.Error(msg, "error", err)
		return models.UserContactInfo{}, errors.New(msg)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := "error reading response body"
		log.Error(msg, "error", err)
		return models.UserContactInfo{}, errors.New(msg)
	}

	var info models.UserContactInfo
	err = json.Unmarshal(body, &info)
	if err != nil {
		msg := "error unmarshaling JSON"
		log.Error(msg, "error", err)
		return models.UserContactInfo{}, errors.New(msg)
	}

	return info, nil
}

var _ ports.UserService = (*UserServiceClient)(nil) // Ensure interface is implemented at compile time
