package chat

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// Based on Go's reference time
const timestampLayout = "02-Jan-06 15:04"

// overwritable for testing
var Timestamp func() string = _timestamp

func _timestamp() string {
	return time.Now().UTC().Format(timestampLayout)
}

type ChatManager struct {
	nameToConn map[string]net.Conn
	chatLog    io.Writer
	mu         sync.Mutex
}

func NewChatManager(chatLog io.Writer) *ChatManager {
	return &ChatManager{
		map[string]net.Conn{},
		chatLog,
		sync.Mutex{}}
}

func (c *ChatManager) Join(name string, conn net.Conn) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.nameToConn[name]; ok {
		return errors.New(fmt.Sprintf(
			"Another \"%s\" is already connected", name))
	}
	c.nameToConn[name] = conn
	log.Printf("%s has joined", name)
	c.broadcast([]byte(fmt.Sprintf(
		"%s * %s %s\n", Timestamp(), name, "has joined")))
	return nil
}

func (c *ChatManager) Quit(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.nameToConn, name)
	log.Printf("%s has quit", name)
	c.broadcast([]byte(fmt.Sprintf(
		"%s * %s %s\n", Timestamp(), name, "has quit")))
}

func (c *ChatManager) broadcast(msg []byte) {
	log.Printf("Broadcasting: %s", string(msg))
	if c.chatLog != nil {
		_, err := c.chatLog.Write(msg)
		if err != nil {
			log.Printf("Error writing to chat log file: %s", err)
		}
	}
	for _, conn := range c.nameToConn {
		go conn.Write(msg)
	}
}

func (c *ChatManager) Broadcast(name string, msg []byte) {
	out := []byte(fmt.Sprintf("%s <%s> %s", Timestamp(), name, msg))
	c.mu.Lock()
	defer c.mu.Unlock()
	c.broadcast(out)
}
