package chat

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

type ChatManager struct {
	nameToConn map[string]net.Conn
	mu         sync.Mutex
}

func NewChatManager() *ChatManager {
	return &ChatManager{
		map[string]net.Conn{},
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
	c.broadcast([]byte(fmt.Sprintf("* %s %s\n", name, "has joined")))
	return nil
}

func (c *ChatManager) Quit(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.nameToConn, name)
	log.Printf("%s has quit", name)
	c.broadcast([]byte(fmt.Sprintf("* %s %s\n", name, "has quit")))
}

func (c *ChatManager) broadcast(msg []byte) {
	for _, conn := range c.nameToConn {
		go conn.Write(msg)
	}
}

func (c *ChatManager) Broadcast(msg string) {
	out := []byte(msg)
	log.Printf("Broadcasting: %s", string(out))
	c.mu.Lock()
	defer c.mu.Unlock()
	c.broadcast(out)
}
