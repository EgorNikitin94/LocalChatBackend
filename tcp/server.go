package tcp

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
	"time"
)

const (
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

type Server struct {
	conns map[string]*Context
	mu    sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		conns: make(map[string]*Context),
		mu:    sync.RWMutex{},
	}
}

func (s *Server) Start(serverType string, host string, port string) {
	addr, err := net.ResolveTCPAddr(serverType, host+":"+port)
	s.handleError(err)

	tcpListener, err := net.ListenTCP(serverType, addr)
	s.handleError(err)
	defer tcpListener.Close()

	fmt.Println("Listening on " + host + ":" + port)

	for {
		c, err := tcpListener.AcceptTCP()
		s.handleError(err)

		conn := s.newContext(c)

		go s.handleConnection(*conn)
	}
}

func (s *Server) handleConnection(c Context) {
	defer c.Close()

	for {
		headerBuffer := make([]byte, HeaderLength)
		if _, errRedHeader := c.Conn().Read(headerBuffer); errRedHeader != nil {
			log.Println(errRedHeader)
			return
		}

		bodySize, _ := binary.Uvarint(headerBuffer)
		requestBuffer := make([]byte, bodySize)
		if _, errReadReq := c.Conn().Read(requestBuffer); errReadReq != nil {
			log.Println(errReadReq)
			return
		}

		req := &pb.Request{}
		proto.Unmarshal(requestBuffer, req)
		fmt.Println(req.GetPayload())

		c.HandleRequest(req)
	}
}

func (s *Server) newContext(conn *net.TCPConn) *Context {
	fmt.Println("New connection " + conn.RemoteAddr().String())

	var c Context = &contextImp{
		id:        uuid.New().String(),
		conn:      conn,
		createdAt: time.Now(),
	}
	s.mu.Lock()
	s.conns[c.ID()] = &c
	s.mu.Unlock()
	return &c
}

func (s *Server) getConn(cid string) (*Context, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if c, ok := s.conns[cid]; ok {
		return c, nil
	}
	return nil, ErrUndefinedConn
}

func (s *Server) NumConns() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.conns)
}

func (s *Server) handleError(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
