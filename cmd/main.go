package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	".."
	// "github.intuit.com/payments/qbchain.git"
)

func main() {
	serverPort := flag.String("port", "8000", "http port number where server will run")
	udpPort := flag.String("udp_port", "10001", "UDP port number where server will listen")
	flag.Parse()

	db, _ := qbchain.MakeDB()
	nodeID := strings.Replace(qbchain.PseudoUUID(), "-", "", -1)
	log.Printf("Starting QB Chain HTTP API Server. Listening at port %q", *serverPort)
	port, _ := strconv.Atoi(*udpPort)
	go qbchain.ListenUDP(port)

	http.Handle("/", qbchain.NewHandler(nodeID, db))
	http.ListenAndServe(fmt.Sprintf(":%s", *serverPort), nil)
}
