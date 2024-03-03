package main

import "net"

type Context struct {
	conn      *net.TCPConn
	sessionId string
	pts       uint32
}
