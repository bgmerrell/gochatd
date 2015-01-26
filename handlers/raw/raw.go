package raw

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/bgmerrell/gochatd/chat"
)

const (
	namePrompt = "What's your name?: "
)

type rawHandler struct {
	buf         []byte
	maxNameSize int
}

func NewRawHandler(bufSize int, maxNameSize int) *rawHandler {
	return &rawHandler{
		make([]byte, bufSize),
		maxNameSize,
	}
}

func (r *rawHandler) validateName(name string) (ok bool) {
	if len(name) == 0 || len(name) > r.maxNameSize {
		return false
	}
	return true
}

func (r *rawHandler) getName(conn net.Conn) (name string, err error) {
	_, err = conn.Write([]byte(namePrompt))
	if err != nil {
		log.Println("Error requesting name:", err.Error())
		return name, errors.New("Error requesting name: " + err.Error())
	}
	n, err := conn.Read(r.buf)
	if err != nil {
		log.Println("Error reading name: ", err.Error())
		return name, errors.New("Error reading name: " + err.Error())
	}
	name = string(bytes.TrimSpace(r.buf[:n]))
	if !r.validateName(name) {
		log.Printf("Invalid name: %s", name)
		return name, errors.New("Invalid name")
	}
	return name, err
}

func (r *rawHandler) Handle(cm *chat.ChatManager, conn net.Conn) {
	name, err := r.getName(conn)
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
		n, err := conn.Read(r.buf)
		if err != nil {
			log.Println(err)
			cm.Quit(name)
			conn.Close()
			return
		}
		cm.Broadcast(name, r.buf[:n])
	}
}
