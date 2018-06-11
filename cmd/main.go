package main

import (
    "flag"
    "fmt"
    "github.intuit.com/payments/qbchain"
    "log"
    "net/http"
    "strings"
)

func main() {
    serverPort := flag.String("port", "8000", "http port number where server will run")
    flag.Parse()

    blockchain := qbchain.NewBlockchain()
    nodeID := strings.Replace(qbchain.PseudoUUID(), "-", "", -1)

    log.Printf("Starting gochain HTTP Server. Listening at port %q", *serverPort)

    http.Handle("/", gochain.NewHandler(blockchain, nodeID))
    http.ListenAndServe(fmt.Sprintf(":%s", *serverPort), nil)
}
