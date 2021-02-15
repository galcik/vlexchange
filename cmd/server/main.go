package main

import (
	"database/sql"
	"github.com/galcik/vlexchange/internal/api"
	"github.com/galcik/vlexchange/internal/coinmarket"
	"github.com/galcik/vlexchange/internal/datastore"
	"log"
	"os"
)

func main() {
	coinmarket.ApiKey = os.Getenv("COINMARKET_API_KEY")

	db, err := sql.Open("postgres", "host=localhost dbname=vlexchange user=vlexchange password=vlexchange")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	srvStore, err := datastore.NewStore(db)
	if err != nil {
		log.Fatal(err)
	}
	server, err := api.NewServer(srvStore)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Starting server.")
	log.Printf("Coinmarket API_KEY: %q", coinmarket.ApiKey)
	err = server.ListenAndServe(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
