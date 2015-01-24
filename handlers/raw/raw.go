package raw

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/bgmerrell/gochatd/handlers"
)

const (
	namePrompt  = "What's your name?: "
	maxNameSize = 32
)

var buf []byte

func init() {
	buf = make([]byte, handlers.BufSize)
}

func validateName(name string) (ok bool) {
	if len(name) == 0 || len(name) > maxNameSize {
		return false
	}
	return true
}

func getName(conn net.Conn) (name []byte, err error) {
	_, err = conn.Write([]byte(namePrompt))
	if err != nil {
		log.Println("Error requesting name:", err.Error())
		return name, errors.New("Error requesting name: " + err.Error())
	}
	n, err := conn.Read(buf)
	if err != nil {
		log.Println(err)
		return name, errors.New("Error reading name: " + err.Error())
	}
	name = bytes.TrimSpace(buf[:n])
	if !validateName(string(name)) {
		log.Printf("Invalid name: %s", name)
		return name, errors.New("Invalid name")
	}
	return name, err
}

func Handle(conn net.Conn) {
	name, err := getName(conn)
	if err != nil {
		_, _ = conn.Write([]byte(fmt.Sprintf("Disconnecting: %s\n", err)))
		conn.Close()
		return
	}
	log.Printf("%s joined", name)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			conn.Close()
			return
		}
		log.Println(string(buf[:n]))
	}
}
