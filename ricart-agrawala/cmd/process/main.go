package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

func broadcastMessage(message string, pool ProcessPool) {
	for _, process := range pool.neighbors {
		go func(process *Process) {
			msg := message + "; from " + strconv.Itoa(pool.local.id)
			buf := []byte(msg)
			_, err := process.conn.Write(buf)
			if err != nil {
				fmt.Println(msg, err)
			}
		}(process)
	}
}

func handleInput(pool ProcessPool) {
	inputChannel := make(chan string)

	// Read input
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			text, _, _ := reader.ReadLine()
			inputChannel <- string(text)
		}
	}()

	for {
		select {
		case message, valid := <-inputChannel:
			if valid {
				broadcastMessage(message, pool)
			} else {
				fmt.Println("channel closed")
			}
		default:
			time.Sleep(time.Second * 1)
		}
	}
}

func listenForMessages(listener *net.UDPConn) {
	buf := make([]byte, 1024)

	for {
		n, addr, err := listener.ReadFromUDP(buf)
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func main() {
	pool, err := InitProcessPool()
	if err != nil {
		fmt.Println(err)
	}

	defer pool.local.conn.Close()
	for _, conn := range pool.neighbors {
		defer conn.conn.Close()
	}

	go listenForMessages(pool.local.conn)
	handleInput(pool)
}
