package chat

import (
	"bytes"
	"container/ring"
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

// history contains a history of messages in a circular buffer.
type history struct {
	message *ring.Ring
	head    *ring.Ring
	maxSize int
}

// newHistory returns a new history object reference.  The size indicates
// the size of the history in number of lines.
func newHistory(size int) *history {
	r := ring.New(size)
	return &history{r, r, size}
}

// insert inserts a line into the chat history
func (h *history) insert(msg []byte) {
	if h.message.Value != nil {
		h.head = h.message.Next()
	}
	h.message.Value = msg
	h.message = h.message.Next()
}

// messages returns n lines of ordered chat messages from the history
func (h *history) messages(n int) []byte {
	// don't allow requests greated than maxSize
	if n > h.maxSize {
		n = h.maxSize
	}
	msgs := []byte{}
	tmp := h.message.Move(-1 * n)
	if tmp.Value == nil {
		// No messages in the chat
		if h.head.Value == nil {
			return msgs
		}
		tmp = h.head
	}
	// this isn't super efficient, but it should work for our purposes.
	for {
		msgs = append(msgs, tmp.Value.([]byte)...)
		tmp = tmp.Next()
		if tmp == h.message {
			break
		}
	}
	return msgs
}

// ChatManager keeps track of clients connected to the chat service and is
// responsible for communications between them.
type ChatManager struct {
	nameToConn map[string]net.Conn
	chatLog    io.Writer
	history    *history
	mu         sync.Mutex
}

// NewChatManager returns an initialized ChatManager
func NewChatManager(chatLog io.Writer, maxHistoryLines int) *ChatManager {
	return &ChatManager{
		map[string]net.Conn{},
		chatLog,
		newHistory(maxHistoryLines),
		sync.Mutex{}}
}

// Join adds a user to the chat manager and announces the join to all clients.
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

// Quit removes a user from the chat manager and announces the quit to all
// clients.
func (c *ChatManager) Quit(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.nameToConn, name)
	log.Printf("%s has quit", name)
	c.broadcast([]byte(fmt.Sprintf(
		"%s * %s %s\n", Timestamp(), name, "has quit")))
}

// broadcast writes msg to all clients known to the ChatManager (but does not
// lock any shared state; it should only be used if you already hold the
// appropriate locks).
func (c *ChatManager) broadcast(msg []byte) {
	if !bytes.HasSuffix(msg, []byte("\n")) {
		msg = append(msg, '\n')
	}
	log.Printf("Broadcasting: %s", string(msg))
	if c.chatLog != nil {
		_, err := c.chatLog.Write(msg)
		if err != nil {
			log.Printf("Error writing to chat log file: %s", err)
		}
	}
	c.history.insert(msg)
	for _, conn := range c.nameToConn {
		go conn.Write(msg)
	}
}

// Broadcast writes msg to all clients known to the ChatManager.
func (c *ChatManager) Broadcast(name string, msg []byte) {
	out := []byte(fmt.Sprintf("%s <%s> %s", Timestamp(), name, msg))
	c.mu.Lock()
	defer c.mu.Unlock()
	c.broadcast(out)
}

// History returns the specifies number of lines (numLines) from the chat
// history as a slices of bytes.
func (c *ChatManager) History(numLines int) []byte {
	return c.history.messages(numLines)
}
