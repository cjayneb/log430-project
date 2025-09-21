package adapters

import (
	"brokerx/core"
	"brokerx/models"
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	Service core.AuthService
}

func (handler * AuthHandler) Login(writer http.ResponseWriter, request *http.Request) {
	var req models.User
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		http.Error(writer, "badly formed user", http.StatusBadRequest)
		return
	}

	user, e := handler.Service.Authenticate(req.Email, req.Password)
	if e != nil {
		http.Error(writer, "unauthorized: " + e.Error(), http.StatusUnauthorized)
		return
	}

	writer.Write([]byte("Welcome " + user.Email))
}