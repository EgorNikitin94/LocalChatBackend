package main

import (
	"LocalChatBackend/proto/pb"
	"encoding/binary"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"os"
	"sync"
)

const (
	HOST         = "localhost"
	PORT         = "8080"
	TYPE         = "tcp"
	HeaderLength = 5
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

		go handleConnection(conn)
	}
}

func handleConnection(conn *net.TCPConn) {
	defer conn.Close()
	fmt.Println("New connection " + conn.RemoteAddr().String())

	for {
		headerBuffer := make([]byte, HeaderLength)
		fmt.Println("New message from connection:" + conn.RemoteAddr().String())
		if _, errRedHeader := conn.Read(headerBuffer); errRedHeader != nil {
			log.Println(errRedHeader)
			return
		}

		bodySize, _ := binary.Uvarint(headerBuffer)
		requestBuffer := make([]byte, bodySize)
		if _, errReadReq := conn.Read(requestBuffer); errReadReq != nil {
			log.Println(errReadReq)
			return
		}

		req := &pb.Request{}
		proto.Unmarshal(requestBuffer, req)
		fmt.Println(req.GetPayload())

		sysIninted := &pb.SysInited{Pts: 1, SessionId: uuid.NewString()}
		response := &pb.Response{
			Id:      req.GetId(),
			Payload: &pb.Response_SysInited{SysInited: sysIninted},
		}
		out, err := proto.Marshal(response)
		if err != nil {
			log.Fatalln("Failed to encode address book:", err)
		}
		responseHeader := make([]byte, HeaderLength)
		fmt.Println(response.GetPayload())
		binary.PutUvarint(responseHeader, uint64(len(out)))
		responseBytes := append(responseHeader, out...)
		conn.Write(responseBytes)
	}

	return
}

func handleError(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
