package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/viper"

	".."
	// "github.intuit.com/payments/qbchain.git"
)

func main() {
	loadConfig()
	db, _ := qbchain.MakeDB()
	nodeID := strings.Replace(qbchain.PseudoUUID(), "-", "", -1)
	serverPort := viper.GetInt("api_port")
	log.Printf("Starting QB Chain HTTP API Server. Listening at port %d", serverPort)
	go qbchain.ListenUDP(viper.GetInt("udp_port"))

	http.Handle("/", qbchain.NewHandler(nodeID, db))
	http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil)
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Failed to load config file:%s", err))
	}
}
