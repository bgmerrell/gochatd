package main

import (
	"log"
	"net"

	"github.com/bgmerrell/gochatd/handlers/raw"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {

		}
		go raw.Handle(conn)
	}
}
