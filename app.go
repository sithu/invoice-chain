package qbchain

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func StartServer(serverPort int) {
	db, _ := MakeDB()
	nodeID := strings.Replace(PseudoUUID(), "-", "", -1)

	go func() {
		log.Printf("Starting QB Chain HTTP API Server. Listening at port %d", serverPort)
		http.Handle("/", NewHandler(nodeID, db))
		http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil)
	}()
}
