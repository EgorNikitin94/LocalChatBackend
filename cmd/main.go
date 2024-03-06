package main

import (
	"LocalChatBackend/tcp"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)

func main() {
	server := tcp.NewServer()
	server.Start(TYPE, HOST, PORT)
}
