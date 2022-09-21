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
		log.Fatalln("Fatal:", err.Error())
	}
	connection, err := net.ListenUDP(consts.UDPProtocol, addr)
	if err != nil {
		log.Fatalln("Fatal:", err.Error())
	}

	defer connection.Close()

	fmt.Println("Listening on:", connection.LocalAddr().String())

	buf := make([]byte, 1024)

	for {
		n, _, err := connection.ReadFromUDP(buf)
		msg := message.DecodeToMessage(buf[0:n])

		fmt.Println(msg.Type, "by", msg.From)

		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}
