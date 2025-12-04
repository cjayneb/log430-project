package handler_adapters

import (
	"brokerx/user-service/core"
	"brokerx/user-service/util"
	"net/http"
	"strconv"
	"strings"
)

type UserHandler struct {
	Service core.UserService
}

func (handler *UserHandler) GetUserContactInfo(w http.ResponseWriter, r *http.Request) {
	log := util.FromContext(r.Context())

	jwt := r.Header.Get("Authorization")
	if !strings.HasPrefix(jwt, "Bearer ") {
		msg := "missing authorization token"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
		return
	}

	userId := r.URL.Query().Get("userId")
	if userId == "" {
		msg := "missing 'userId' query parameter"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		msg := "invalid 'userId' query parameter"
		log.Warn(msg, "error", err)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}

	user, err := handler.Service.GetUserContactInfo(r.Context(), userIdInt)
	if err != nil {
		msg := "error fetching user"
		log.Error(msg, "error", err)
		util.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: msg})
		return
	}
	util.WriteJSON(w, http.StatusOK, user)
}
