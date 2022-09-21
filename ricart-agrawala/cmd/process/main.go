package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/ricart-agrawala/lib/consts"
)

func parseArguments(args []string) (int, []string) {
	headId, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatalln("Fatal: ", err.Error())
	}

	ports := args[2:]
	numberOfProcesses := len(ports)

	addresses := make([]string, numberOfProcesses)
	for i, port := range ports {
		addresses[i] = consts.LocalIp + port
	}

	return headId, addresses
}

func main() {
	headId, addresses := parseArguments(os.Args)

	head := NewHeadProcess(headId)
	if err := head.InitializeConnections(addresses); err != nil {
		log.Fatalln("Fatal:", err.Error())
	}

	defer head.recv.Close()
	defer head.sharedResource.Close()
	for _, link := range head.links {
		defer link.conn.Close()
	}

	fmt.Println("Listening on:", head.recv.LocalAddr().String())

	go head.ListenForMessages()
	head.ListenForInput()
}
