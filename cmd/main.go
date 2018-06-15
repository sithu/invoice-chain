package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	".."
	// "github.intuit.com/payments/qbchain.git"
)

func main() {
	serverPort := flag.String("port", "8000", "http port number where server will run")
	flag.Parse()

	db, _ := qbchain.MakeDB()
	blockchain := qbchain.NewBlockchain(db)
	nodeID := strings.Replace(qbchain.PseudoUUID(), "-", "", -1)
	log.Printf("Starting QB Chain HTTP Server. Listening at port %q", *serverPort)

	http.Handle("/", qbchain.NewHandler(blockchain, nodeID, db))
	http.ListenAndServe(fmt.Sprintf(":%s", *serverPort), nil)
}
