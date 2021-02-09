package main

import (
	"github.com/galcik/vlexchange/internal/api"
	"github.com/galcik/vlexchange/internal/coinmarket"
	"github.com/galcik/vlexchange/internal/db"
	"log"
	"os"
)

func main() {
	coinmarket.ApiKey = os.Getenv("COINMARKET_API_KEY")
	store, _ := db.NewStore("host=localhost dbname=vlexchange user=vlexchange password=vlexchange")
	server, _ := api.NewServer(store)
	log.Print("Starting BTC exchange server.")
	log.Printf("Coinmarket API_KEY: %q", coinmarket.ApiKey)
	log.Fatal(server.ListenAndServe(":8080"))
}
