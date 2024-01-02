package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

const (
	HOST = "127.0.0.1"
	PORT = "80"
	NET  = "tcp"
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
	addr, err := net.ResolveTCPAddr(NET, HOST+":"+PORT)
	handleError(err)

	tcpListener, err := net.ListenTCP(NET, addr)
	handleError(err)
	defer tcpListener.Close()

	c := &Connections{
		connections: make(map[string]*net.TCPConn),
	}
	// todo create broudcaster and ch

	fmt.Println("Listening on " + HOST + ":" + PORT)

	for {
		conn, err := tcpListener.AcceptTCP()
		handleError(err)
		c.New(conn.RemoteAddr().String(), conn)

		// todo handle request
	}
}

func handleError(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
