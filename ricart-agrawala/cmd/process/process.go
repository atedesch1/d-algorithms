package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/ricart-agrawala/lib/consts"
	"github.com/ricart-agrawala/lib/message"
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

	if err := p.initializeSharedResourceLink(consts.LocalIp + consts.SharedResourcePort); err != nil {
		return err
	}

	return nil
}

func (p *HeadProcess) initializeReceiver(address string) error {
	recvAddr, err := net.ResolveUDPAddr(consts.UDPProtocol, address)
	if err != nil {
		return err
	}
	p.addr = recvAddr

	recv, err := net.ListenUDP(consts.UDPProtocol, recvAddr)
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

		addr, err := net.ResolveUDPAddr(consts.UDPProtocol, address)
		if err != nil {
			return err
		}
		process.addr = addr

		conn, err := net.DialUDP(consts.UDPProtocol, nil, addr)
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
	sharedResourceAddr, err := net.ResolveUDPAddr(consts.UDPProtocol, address)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP(consts.UDPProtocol, nil, sharedResourceAddr)
	if err != nil {
		return err
	}

	p.sharedResource = conn
	return nil
}

func (p *HeadProcess) GetLinkWithId(id int) (*Process, error) {
	for _, link := range p.links {
		if link.id == id {
			return link, nil
		}
	}
	return &Process{}, errors.New("no link with id " + strconv.Itoa(id))
}

func (p *HeadProcess) SendMessage(msg *message.Message, conn *net.UDPConn) error {
	buf := msg.EncodeToBytes()
	_, err := conn.Write(buf)
	if err != nil {
		fmt.Println(msg, err)
	}
	return err
}

func (p *HeadProcess) BroadcastMessage(msg *message.Message) {
	for _, link := range p.links {
		go p.SendMessage(msg, link.conn)
	}
}

func (p *HeadProcess) ListenForMessages() {
	buf := make([]byte, 1024)

	for {
		n, addr, err := p.recv.ReadFromUDP(buf)
		msg := message.DecodeToMessage(buf[0:n])

		fmt.Println("Received ", msg.Content, " from ", addr)

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
		case content, valid := <-inputChannel:
			if valid {
				msg := message.NewMessage(p.id, p.clock, content)
				p.BroadcastMessage(msg)
			} else {
				fmt.Println("channel closed")
			}
		default:
			time.Sleep(time.Second * 1)
		}
	}
}

func (p *HeadProcess) RequestSharedResource() {
}

func (p *HeadProcess) AcquireSharedResource() {
	msg := message.NewMessage(p.id, p.clock, "cs")
	p.SendMessage(msg, p.sharedResource)
}
