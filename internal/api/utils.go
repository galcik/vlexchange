package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

func isValidOrderType(orderType string) bool {
	orderType = strings.ToLower(orderType)
	return orderType == "buy" || orderType == "sell"
}

func writeJSONResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
