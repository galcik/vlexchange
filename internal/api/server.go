package api

import (
	"github.com/galcik/vlexchange/internal/db"
	"github.com/gorilla/mux"
	"net/http"
)

type Server struct {
	store  *db.Store
	router *mux.Router
}

// NewServer creates a new HTTP server and set up routing.
func NewServer(store *db.Store) (*Server, error) {
	server := &Server{
		store: store,
	}
	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	server.router = mux.NewRouter()
	server.router.HandleFunc("/register", server.handleRegister).Methods(http.MethodPost)
	server.router.HandleFunc("/balance", server.handleGetBalance).Methods(http.MethodGet)
	server.router.HandleFunc("/balance", server.handlePostBalance).Methods(http.MethodPost)
	server.router.HandleFunc("/standing_orders", server.handlePostStandingOrder).Methods(http.MethodPost)
	server.router.HandleFunc("/standing_orders/{id:[0-9]+}", server.handlePostStandingOrder).Methods(http.MethodGet)
	server.router.HandleFunc("/standing_orders/{id:[0-9]+}", server.handlePostStandingOrder).Methods(http.MethodDelete)

	// OpenAPI
	fs := http.FileServer(http.Dir("./openapi/swaggerui"))
	server.router.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", fs))
	server.router.HandleFunc(
		"/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./openapi/openapi.yaml")
		},
	)
}

func (server *Server) ListenAndServe(addr string) error {
	httpServer := &http.Server{Addr: addr, Handler: server.router}
	return httpServer.ListenAndServe()
}
