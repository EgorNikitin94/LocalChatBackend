package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)

type Connections struct {
	connections map[string]*net.TCPConn
	sync.RWMutex
}

// New connection
func (c *Connections) New(key string, conn *net.TCPConn) bool {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.connections[key]; !ok {
		return true
	}
	return false
}

// Delete connection
func (c *Connections) Delete(key string) bool {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.connections[key]; !ok {
		delete(c.connections, key)
		return true
	}
	return false
}

// Read connection
func (c *Connections) Read(key string) *net.TCPConn {
	c.RLock()
	defer c.RUnlock()

	if connection, ok := c.connections[key]; ok {
		return connection
	}
	return nil
}

func main() {
	addr, err := net.ResolveTCPAddr(TYPE, HOST+":"+PORT)
	handleError(err)

	tcpListener, err := net.ListenTCP(TYPE, addr)
	handleError(err)
	defer tcpListener.Close()

	var broadcaster = make(chan byte)
	defer close(broadcaster)

	c := &Connections{
		connections: make(map[string]*net.TCPConn),
	}

	fmt.Println("Listening on " + HOST + ":" + PORT)

	for {
		conn, err := tcpListener.AcceptTCP()
		handleError(err)
		c.New(conn.RemoteAddr().String(), conn)

		handleRequest(conn)
	}
}

func handleRequest(conn *net.TCPConn) {
	fmt.Println("New connection " + conn.RemoteAddr().String())
}

func handleError(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
