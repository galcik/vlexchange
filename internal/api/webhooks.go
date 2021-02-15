package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/galcik/vlexchange/internal/datastore/queries"
	"log"
	"net/http"
)

func (server *Server) callWebhooks(orderIds []int32) {
	store := server.store.WithContext(context.Background())
	orders, err := store.GetStandingOrders(orderIds)
	if err != nil {
		return
	}

	for _, order := range orders {
		callOrderChangedWebhook(order)
	}
}

func callOrderChangedWebhook(order queries.StandingOrder) {
	if !order.WebhookUrl.Valid {
		return
	}

	values := map[string]interface{}{"orderId": order.ID}

	callbackBody, err := json.Marshal(values)
	if err != nil {
		return
	}

	client := &http.Client{}
	resp, err := client.Post(order.WebhookUrl.String, "application/json", bytes.NewBuffer(callbackBody))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	log.Printf("callback to %q done", order.WebhookUrl.String)
}
