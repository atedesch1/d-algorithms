package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type Process struct {
	id   int
	addr *net.UDPAddr
	conn *net.UDPConn
}

type HeadProcess struct {
	id             int
	clock          int
	addr           *net.UDPAddr
	recv           *net.UDPConn
	links          []*Process
	sharedResource *net.UDPConn
}

func NewHeadProcess(id int) *HeadProcess {
	return &HeadProcess{id: id, clock: 0}
}

func (p *HeadProcess) InitializeConnections(addresses []string) error {
	if err := p.initializeReceiver(addresses[p.id-1]); err != nil {
		return err
	}

	if err := p.initializeProcessLinks(addresses); err != nil {
		return err
	}

	if err := p.initializeSharedResourceLink(ip + sharedResourcePort); err != nil {
		return err
	}

	return nil
}

func (p *HeadProcess) initializeReceiver(address string) error {
	recvAddr, err := net.ResolveUDPAddr(protocol, address)
	if err != nil {
		return err
	}
	p.addr = recvAddr

	recv, err := net.ListenUDP(protocol, recvAddr)
	if err != nil {
		return err
	}
	p.recv = recv

	return nil
}

func (p *HeadProcess) initializeProcessLinks(addresses []string) error {
	var links []*Process

	for i, address := range addresses {
		procId := i + 1
		if procId == p.id {
			continue
		}

		process := &Process{
			id: procId,
		}

		addr, err := net.ResolveUDPAddr(protocol, address)
		if err != nil {
			return err
		}
		process.addr = addr

		conn, err := net.DialUDP(protocol, nil, addr)
		if err != nil {
			return err
		}
		process.conn = conn

		links = append(links, process)
	}

	p.links = links
	return nil
}

func (p *HeadProcess) initializeSharedResourceLink(address string) error {
	sharedResourceAddr, err := net.ResolveUDPAddr(protocol, address)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP(protocol, nil, sharedResourceAddr)
	if err != nil {
		return err
	}

	p.sharedResource = conn
	return nil
}

func (p *HeadProcess) BroadcastMessage(message string) {
	for _, link := range p.links {
		go func(link *Process) {
			msg := message + "; from " + strconv.Itoa(p.id)
			buf := []byte(msg)

			_, err := link.conn.Write(buf)
			if err != nil {
				fmt.Println(msg, err)
			}
		}(link)
	}
}

func (p *HeadProcess) ListenForMessages() {
	buf := make([]byte, 1024)

	for {
		n, addr, err := p.recv.ReadFromUDP(buf)
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func (p *HeadProcess) ListenForInput() {
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
				p.BroadcastMessage(message)
			} else {
				fmt.Println("channel closed")
			}
		default:
			time.Sleep(time.Second * 1)
		}
	}
}
