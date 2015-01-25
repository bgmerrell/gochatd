package raw

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/bgmerrell/gochatd/chat"
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

func getName(conn net.Conn) (name string, err error) {
	_, err = conn.Write([]byte(namePrompt))
	if err != nil {
		log.Println("Error requesting name:", err.Error())
		return name, errors.New("Error requesting name: " + err.Error())
	}
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("Error reading name: ", err.Error())
		return name, errors.New("Error reading name: " + err.Error())
	}
	name = string(bytes.TrimSpace(buf[:n]))
	if !validateName(name) {
		log.Printf("Invalid name: %s", name)
		return name, errors.New("Invalid name")
	}
	return name, err
}

func Handle(cm *chat.ChatManager, conn net.Conn) {
	name, err := getName(conn)
	if err != nil {
		_, _ = conn.Write([]byte(fmt.Sprintf("Disconnecting: %s\n", err)))
		conn.Close()
		return
	}
	err = cm.Join(name, conn)
	if err != nil {
		_, _ = conn.Write([]byte(fmt.Sprintf("Disconnecting: %s\n", err)))
		conn.Close()
		return
	}
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			cm.Quit(name)
			conn.Close()
			return
		}
		cm.Broadcast(fmt.Sprintf("<%s> %s", name, string(buf[:n])))
	}
}
