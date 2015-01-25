package main

import (
	"log"
	"net"

	"github.com/bgmerrell/gochatd/chat"
	"github.com/bgmerrell/gochatd/handlers/raw"
)

func main() {
	cm := chat.NewChatManager()
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {

		}
		go raw.Handle(cm, conn)
	}
}
