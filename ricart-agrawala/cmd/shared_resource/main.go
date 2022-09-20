package main

import "net"

func main() {
	address, _ := net.ResolveUDPAddr("udp", ":10001")
	connection, _ := net.ListenUDP("udp", address)

	defer connection.Close()

	for {
	}
}
