package qbchain

import (
	"log"
	"net"
	"os"
	"strconv"
)

func SendUDP(data []byte, udpServer string) error {
	conn, err := net.Dial("udp", udpServer)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer conn.Close()
	conn.Write([]byte("Hello from client"))
	return nil
}

// listen to incoming udp packets
func ListenUDP(port int) {
	/* Lets prepare a address at any address at port 10001*/
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	log.Println("UDP Server is listening at port " + strconv.Itoa(port) + "...")
	defer ServerConn.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		log.Println("Received ", string(buf[0:n]), " from ", addr)

		if err != nil {
			log.Println("Error: ", err)
		}
	}
}

/* A Simple function to verify error */
func CheckError(err error) {
	if err != nil {
		log.Println("Error: ", err)
		os.Exit(0)
	}
}
