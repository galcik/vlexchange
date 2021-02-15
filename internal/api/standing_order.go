package api

import (
	"encoding/json"
	"github.com/galcik/vlexchange/internal/currency"
	"github.com/galcik/vlexchange/internal/datastore"
	"github.com/galcik/vlexchange/internal/datastore/queries"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

type postStandingOrderRequest struct {
	Quantity   string `json:"quantity"`
	Type       string `json:"type"`
	LimitPrice string `json:"limitPrice"`
	WebhookUrl string `json:"webhookUrl"`
}

type postStandingOrderResponse struct {
	OrderId int32 `json:"orderId"`
}

func (server *Server) handlePostStandingOrder(w http.ResponseWriter, req *http.Request) {
	store := server.store.WithContext(req.Context())
	account, err := store.GetAccountByToken(req.Header.Get("X-Token"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if account == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}

	var payload postStandingOrderRequest
	err = json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !isValidOrderType(payload.Type) {
		http.Error(w, "malformed order type", http.StatusBadRequest)
		return
	}
	orderType := queries.OrderType(strings.ToLower(payload.Type))
	quantity, err := currency.ParseBTC(payload.Quantity)
	if err != nil {
		http.Error(w, "malformed quantity", http.StatusBadRequest)
		return
	}
	limitPrice, err := currency.ParseUSD(payload.Quantity)
	if err != nil {
		http.Error(w, "malformed limitPrice", http.StatusBadRequest)
		return
	}

	if quantity <= 0 {
		http.Error(w, "no quantity", http.StatusBadRequest)
		return
	}

	if limitPrice < 0 {
		http.Error(w, "negative quantity", http.StatusBadRequest)
		return
	}

	standingOrder, affectedOrderIds, err := store.CreateStandingOrder(
		datastore.CreateStandingOrderParams{
			AccountID:  account.ID,
			OrderType:  orderType,
			Quantity:   quantity,
			LimitPrice: limitPrice,
		},
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go server.callWebhooks(affectedOrderIds)

	writeJSONResponse(w, postStandingOrderResponse{OrderId: standingOrder.ID})
}

type getStandingOrderResponse struct {
	ID             int32  `json:"id"`
	Type           string `json:"type"`
	State          string `json:"state"`
	Quantity       string `json:"quantity"`
	FilledQuantity string `json:"filledQuantity"`
	LimitPrice     string `json:"limitPrice"`
	AvgPrice       string `json:"avgPrice"`
}

func (server *Server) handleGetStandingOrder(w http.ResponseWriter, req *http.Request) {
	orderId, _ := strconv.Atoi(mux.Vars(req)["id"])
	store := server.store.WithContext(req.Context())

	order, err := store.GetStandingOrder(int32(orderId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if order == nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	account, err := store.GetAccount(order.AccountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if account == nil || account.Token != req.Header.Get("X-Token") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	writeJSONResponse(
		w,
		getStandingOrderResponse{
			ID:             order.ID,
			Type:           strings.ToUpper(string(order.Type)),
			State:          strings.ToUpper(string(order.State)),
			Quantity:       currency.BTC(order.Quantity).String(),
			FilledQuantity: currency.BTC(order.FilledQuantity).String(),
			LimitPrice:     currency.USD(order.LimitPrice).String(),
			AvgPrice:       "0",
		},
	)
}

func (server *Server) handleDeleteStandingOrder(w http.ResponseWriter, req *http.Request) {
	orderId, _ := strconv.Atoi(mux.Vars(req)["id"])
	store := server.store.WithContext(req.Context())

	order, err := store.GetStandingOrder(int32(orderId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if order == nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	account, err := store.GetAccount(order.AccountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if account == nil || account.Token != req.Header.Get("X-Token") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err = store.DeleteStandingOrder(int32(orderId)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]bool{"success": true})
}
