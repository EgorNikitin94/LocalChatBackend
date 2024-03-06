package tcp

import (
	"LocalChatBackend/proto/pb"
	"encoding/binary"
	"fmt"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"sync"
	"time"
)

type Context interface {
	ID() string
	PTS() uint32
	Conn() *net.TCPConn
	RemoteAddr() net.Addr
	CreatedAt() time.Time
	Close() error
	Send(resp *pb.Response) error
	HandleRequest(req *pb.Request)
}

type contextImp struct {
	id        string
	conn      *net.TCPConn
	pts       uint32
	mu        sync.RWMutex
	createdAt time.Time
}

func (c *contextImp) ID() string {
	return c.id
}

func (c *contextImp) PTS() uint32 {
	return c.pts
}

func (c *contextImp) Conn() *net.TCPConn {
	return c.conn
}

func (c *contextImp) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *contextImp) CreatedAt() time.Time {
	return c.createdAt
}

func (c *contextImp) HandleRequest(req *pb.Request) {
	payload := req.GetPayload()
	switch payload.(type) {
	case *pb.Request_SysInit:
		sysInited := &pb.SysInited{Pts: c.PTS(), SessionId: c.ID()}
		response := &pb.Response{
			Id:      req.GetId(),
			Payload: &pb.Response_SysInited{SysInited: sysInited},
		}
		c.Send(response)
	case *pb.Request_Ping:
		id := req.GetPing().GetId()
		pong := &pb.Pong{Id: id}
		response := &pb.Response{
			Id:      req.GetId(),
			Payload: &pb.Response_Pong{Pong: pong},
		}
		c.Send(response)
	case *pb.Request_CheckLogin:
		break
	case *pb.Request_SendSmsCode:
		break
	case *pb.Request_SighIn:
		break
	case *pb.Request_SignUp:
		break
	case *pb.Request_Authorize:
		break
	default:
		pbError := &pb.SysError{
			Code:        500,
			Reason:      "Unsupported Request",
			Description: "This Request unsupport now",
		}
		response := &pb.Response{
			Id:      req.GetId(),
			Payload: &pb.Response_Error{Error: pbError},
		}
		c.Send(response)
	}
}

func (c *contextImp) Send(resp *pb.Response) error {
	out, err := proto.Marshal(resp)
	if err != nil {
		log.Fatalln("Failed to encode address book:", err)
		return err
	}
	responseHeader := make([]byte, HeaderLength)
	fmt.Println(resp.GetPayload())
	binary.PutUvarint(responseHeader, uint64(len(out)))
	responseBytes := append(responseHeader, out...)
	_, errWrite := c.conn.Write(responseBytes)
	return errWrite
}

func (c *contextImp) Close() error {
	return c.conn.Close()
}
