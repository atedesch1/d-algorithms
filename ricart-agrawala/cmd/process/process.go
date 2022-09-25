package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ricart-agrawala/lib/consts"
	"github.com/ricart-agrawala/lib/message"
)

type Process struct {
	id   int
	addr *net.UDPAddr
	conn *net.UDPConn
}

type State string

const (
	Wanted   State = "WANTED"
	Held     State = "HELD"
	Released State = "RELEASED"
)

type MutexWrap struct {
	item  interface{}
	mutex sync.Mutex
}

type HeadProcess struct {
	id    int
	clock MutexWrap
	state State

	addr *net.UDPAddr
	recv *net.UDPConn

	links []*Process

	sharedResource *net.UDPConn

	replyQueue MutexWrap
	responded  MutexWrap
}

func NewHeadProcess(id int) *HeadProcess {
	return &HeadProcess{
		id:         id,
		clock:      MutexWrap{item: 0, mutex: sync.Mutex{}},
		state:      Released,
		responded:  MutexWrap{item: 0, mutex: sync.Mutex{}},
		replyQueue: MutexWrap{item: make([]int, 0), mutex: sync.Mutex{}},
	}
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
	return err
}

func (p *HeadProcess) ListenForMessages() {
	buf := make([]byte, 1024)

	for {
		n, _, err := p.recv.ReadFromUDP(buf)
		msg := message.DecodeToMessage(buf[0:n])

		go p.handleMessage(msg)

		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}

func (p *HeadProcess) ListenForInput() {
	inputChannel := make(chan string)

	// Read input
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			input, _, err := reader.ReadLine()
			if err != nil {
				fmt.Println("Error:", err.Error())
			}
			inputChannel <- string(input)
		}
	}()

	for input := range inputChannel {
		if input == "x" {
			if err := p.requestSharedResource(); err != nil {
				fmt.Println("Error:", err.Error())
			}
		} else if input == strconv.Itoa(p.id) {
			p.incrementClock(1)
		} else {
			fmt.Println("Invalid input")
		}
	}
}

func (p *HeadProcess) alterState(nextState State) {
	p.state = nextState
	fmt.Println("State:", p.state+"; Clock:", p.clock.item.(int))
}

func (p *HeadProcess) incrementClock(increment int) {
	p.clock.mutex.Lock()
	p.clock.item = p.clock.item.(int) + increment
	p.clock.mutex.Unlock()
}

func (p *HeadProcess) matchAndincrementClock(timestamp int) {
	increment := 1
	if dif := timestamp - p.clock.item.(int); dif > 0 {
		increment += dif
	}
	p.incrementClock(increment)
}

func (p *HeadProcess) requestSharedResource() error {
	if p.state != Released {
		return errors.New("input ignored, process not in released state")
	}

	p.incrementClock(1)
	p.alterState(Wanted)
	msg := message.NewMessage(p.id, p.clock.item.(int), message.Request)
	for _, link := range p.links {
		go p.SendMessage(msg, link.conn)
	}
	return nil
}

func (p *HeadProcess) replyToRequest(recipientId int) error {
	reply := message.NewMessage(p.id, p.clock.item.(int), message.Reply)
	link, err := p.GetLinkWithId(recipientId)
	if err != nil {
		return err
	}
	go p.SendMessage(reply, link.conn)
	return nil
}

func (p *HeadProcess) acquireSharedResource() {
	p.alterState(Held)
	msg := message.NewMessage(p.id, p.clock.item.(int), message.Acquire)
	go p.SendMessage(msg, p.sharedResource)

	fmt.Println("Acquired CS; Clock:", p.clock.item.(int))
	time.Sleep(consts.CSTimeout)
}

func (p *HeadProcess) releaseSharedResource() error {
	p.incrementClock(1)
	fmt.Println("Released CS; Clock:", p.clock.item.(int))
	p.alterState(Released)
	for _, id := range p.replyQueue.item.([]int) {
		if err := p.replyToRequest(id); err != nil {
			return err
		}
	}
	p.replyQueue.mutex.Lock()
	p.replyQueue.item = make([]int, 0)
	p.replyQueue.mutex.Unlock()
	return nil
}

func (p *HeadProcess) handleMessage(msg *message.Message) error {
	switch msg.Type {
	case message.Request:
		if p.state == Held || (p.state == Wanted && p.clock.item.(int) < msg.Timestamp) {
			p.replyQueue.mutex.Lock()
			p.replyQueue.item = append(p.replyQueue.item.([]int), msg.From)
			p.replyQueue.mutex.Unlock()
		} else if err := p.replyToRequest(msg.From); err != nil {
			return err
		}
	case message.Reply:
		p.responded.mutex.Lock()
		p.responded.item = p.responded.item.(int) + 1
		p.responded.mutex.Unlock()
	}

	p.matchAndincrementClock(msg.Timestamp)

	if p.responded.item == len(p.links) { // Received replies from every process
		p.responded.mutex.Lock()
		p.responded.item = 0
		p.responded.mutex.Unlock()

		p.acquireSharedResource()

		if err := p.releaseSharedResource(); err != nil {
			return err
		}
	}

	return nil
}
