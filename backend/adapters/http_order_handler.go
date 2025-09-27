package adapters

import (
	"brokerx/models"
	"brokerx/ports"
	"html/template"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/schema"
)

const ORDER_PAGE = "order.html"

type OrderViewData struct {
    Email   string
    Orders  []*models.Order
    Error   string
    Success string
}

type OrderHandler struct {
	Service ports.OrderService
	FrontendPath string
}

func (handler *OrderHandler) PlaceOrder(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	order, decodeErr := validateOrderForm(request)
	if err != nil || decodeErr != nil {
		var errorMsg string
		if err != nil {
			errorMsg = err.Error()
		} else {
			errorMsg = decodeErr.Error()
		}
		log.Errorf("Client error when placing order : %v", errorMsg)
		writer.WriteHeader(http.StatusBadRequest)
		handler.renderTemplate(writer, request, ORDER_PAGE, OrderViewData{Error: errorMsg})
		return
	}

	err = handler.Service.PlaceOrder(order)
	if err != nil {
		log.Errorf("Internal error when placing order : %v", err)
		writer.WriteHeader(http.StatusInternalServerError)
		handler.renderTemplate(writer, request, ORDER_PAGE, OrderViewData{Error: err.Error()})
		return
	}

	writer.WriteHeader(http.StatusCreated)
	handler.renderTemplate(writer, request, ORDER_PAGE, OrderViewData{Success: "order placed sucessfully!"})
}

func (handler *OrderHandler) GetOrders(writer http.ResponseWriter, request *http.Request) {
	handler.renderTemplate(writer, request, ORDER_PAGE, OrderViewData{})
}

func (handler *OrderHandler) renderTemplate(w http.ResponseWriter, r *http.Request, name string, data OrderViewData) {
    tpl, err := template.ParseFiles(handler.FrontendPath+"/templates/base.html", handler.FrontendPath+"/templates/"+name)
    if err != nil {
        http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
        return
    }

	userId := r.Context().Value(USER_ID_KEY).(string)
	userEmail := r.Context().Value(USER_EMAIL_KEY).(string)

    orders, err := handler.Service.GetOrdersForUser(userId)
    if err != nil {
        log.Errorf("Failed to fetch orders: %v", err)
        orders = []*models.Order{}
    }

	data.Email = userEmail
	data.Orders = orders

	err = tpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
	}
}

func validateOrderForm(request *http.Request) (*models.Order, error) {
	var order models.Order
	decoder := schema.NewDecoder()
	err := decoder.Decode(&order, request.PostForm);
	order.UserID = request.Context().Value(USER_ID_KEY).(string)
	
	if err != nil || !isValidOrder(&order) {
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
        order.Timing != ""
}