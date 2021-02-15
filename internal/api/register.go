package api

import (
	"context"
	"encoding/json"
	"github.com/galcik/vlexchange/internal/datastore/queries"
	"github.com/google/uuid"
	"net/http"
)

type registerRequest struct {
	Username string `json:"username"`
}

type registerResponse struct {
	Token string `json:"token"`
}

func (server *Server) handleRegister(w http.ResponseWriter, req *http.Request) {
	store := server.store.WithContext(req.Context())
	var payload registerRequest

	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := uuid.NewRandom()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var newAccount queries.Account
	err = store.ExecuteTx(
		func(ctx context.Context, q queries.Querier) error {
			newAccount, err = q.CreateAccount(
				ctx,
				queries.CreateAccountParams{Username: payload.Username, Token: token.String()},
			)
			return err
		},
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, registerResponse{Token: newAccount.Token})
}
