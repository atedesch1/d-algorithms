package main

import (
	"fmt"
	"net"
)

var protocol string = "udp"
var address string = "127.0.0.1:10000"

func main() {
	addr, _ := net.ResolveUDPAddr(protocol, address)
	connection, _ := net.ListenUDP(protocol, addr)

	defer connection.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, err := connection.ReadFromUDP(buf)
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}
