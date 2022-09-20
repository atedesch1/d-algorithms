package main

import (
	"net"
	"os"
	"strconv"
)

var protocol string = "udp"
var ip string = "127.0.0.1"

type Process struct {
	id   int
	conn *net.UDPConn
}

type ProcessPool struct {
	local     *Process
	neighbors []*Process
}

func registerProcess(id int, address string) (*Process, error) {
	var process *Process

	addr, err := net.ResolveUDPAddr(protocol, address)
	if err != nil {
		return process, err
	}
	conn, err := net.DialUDP(protocol, nil, addr)
	if err != nil {
		return process, err
	}

	process = &Process{
		id:   id,
		conn: conn,
	}

	return process, nil
}

func InitProcessPool() (ProcessPool, error) {
	var pool ProcessPool

	id, err := strconv.Atoi(os.Args[1])
	if err != nil {
		return pool, err
	}
	ports := os.Args[2:]
	localPort := ports[id-1]
	localAddrStr := ip + localPort

	// Register local process
	localAddr, err := net.ResolveUDPAddr(protocol, localAddrStr)
	if err != nil {
		return pool, err
	}
	conn, err := net.ListenUDP(protocol, localAddr)
	if err != nil {
		return pool, err
	}
	pool.local = &Process{
		id:   id,
		conn: conn,
	}

	// Register neighbor processes links
	for i, port := range ports {
		procId := i + 1
		procAddress := ip + port

		if procId == id {
			continue
		}

		process, err := registerProcess(procId, procAddress)
		if err != nil {
			return pool, err
		}

		pool.neighbors = append(pool.neighbors, process)
	}

	return pool, err
}
