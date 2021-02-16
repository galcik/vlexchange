package api

import (
	"encoding/json"
	"fmt"
	"github.com/galcik/vlexchange/internal/currency"
	"net/http"
	"strings"
)

type getBalanceResponse struct {
	BTC           string `json:"BTC"`
	USD           string `json:"USD"`
	USDEquivalent string `json:"USDEquivalent"`
}

func (server *Server) handleGetBalance(w http.ResponseWriter, req *http.Request) {
	store := server.store.WithContext(req.Context())
	account, err := store.GetAccountByToken(req.Header.Get("X-Token"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if account == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}

	btcAmount := currency.BTC(account.BtcAmount)
	usdAmount := currency.USD(account.UsdAmount)

	btcPrice, err := server.coinmarketService.GetBTCPriceInUSD(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(
		w, getBalanceResponse{
			BTC:           btcAmount.String(),
			USD:           usdAmount.String(),
			USDEquivalent: btcAmount.USD(btcPrice).String(),
		},
	)
}

type postBalanceRequest struct {
	TopupAmount string `json:"topupAmount"`
	Currency    string `json:"currency"`
}

type postBalanceResponse struct {
	Success bool `json:"success"`
}

func (server *Server) handlePostBalance(w http.ResponseWriter, req *http.Request) {
	store := server.store.WithContext(req.Context())
	account, err := store.GetAccountByToken(req.Header.Get("X-Token"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if account == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}

	var payload postBalanceRequest
	err = json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	payload.Currency = strings.ToUpper(payload.Currency)
	usdAmount := currency.USD(0)
	btcAmount := currency.BTC(0)
	switch payload.Currency {
	case "USD":
		if usdAmount, err = currency.ParseUSD(payload.TopupAmount); err != nil {
			http.Error(w, "invalid amount", http.StatusBadRequest)
			return
		}
	case "BTC":
		if btcAmount, err = currency.ParseBTC(payload.TopupAmount); err != nil {
			http.Error(w, "invalid amount", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, fmt.Sprintf("unsupported currency %q", payload.Currency), http.StatusBadRequest)
		return
	}

	success, err := store.DepositAccount(account.ID, btcAmount, usdAmount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, postBalanceResponse{Success: success})
}
