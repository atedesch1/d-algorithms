package main

import (
	"fmt"
	"log"
	"net"

	"github.com/ricart-agrawala/lib/consts"
	"github.com/ricart-agrawala/lib/message"
)

func main() {
	addr, err := net.ResolveUDPAddr(consts.UDPProtocol, consts.LocalIp+consts.SharedResourcePort)
	if err != nil {
		log.Fatalln("Fatal: ", err.Error())
	}
	connection, err := net.ListenUDP(consts.UDPProtocol, addr)
	if err != nil {
		log.Fatalln("Fatal: ", err.Error())
	}

	defer connection.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, err := connection.ReadFromUDP(buf)
		msg := message.DecodeToMessage(buf[0:n])

		fmt.Println("Received ", msg.Content, " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}
