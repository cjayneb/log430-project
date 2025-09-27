package adapters

import (
	"brokerx/models"
	"brokerx/ports"
	"log"
	"net/http"

	"github.com/gorilla/schema"
)

type OrderHandler struct {
	Service ports.OrderService
}

func (handler *OrderHandler) PlaceOrder(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	order, decodeErr := validateOrderForm(request)
	if err != nil || decodeErr != nil {
		writer.WriteHeader(http.StatusBadRequest)
		http.ServeFile(writer, request, "./frontend/order_failed.html")
		return
	}

	err = handler.Service.PlaceOrder(order)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		http.ServeFile(writer, request, "./frontend/order_failed.html")
		return
	}

	writer.WriteHeader(http.StatusCreated)
	http.ServeFile(writer, request, "./frontend/order_created.html")
}

func validateOrderForm(request *http.Request) (*models.Order, error) {
	var order models.Order
	decoder := schema.NewDecoder()
	if err := decoder.Decode(&order, request.PostForm); err != nil || !isValidOrder(&order) {
		return nil, err
	}
	return &order, nil
}

func isValidOrder(order *models.Order) bool {
	log.Printf("Validating order: %+v", order)
    return order.UserID != "" &&
        order.Symbol != "" &&
        order.Type != "" &&
        order.Action != "" &&
        order.Quantity > 0 &&
        order.UnitPrice > 0 &&
        order.Timing != "" &&
        order.Status != ""
}