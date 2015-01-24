package dummyconn

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bgmerrell/gochatd/handlers"
)

// dummyConn implements net.Conn.  It is meant to use as a mock connection
// for unit tests.  It communicates reads and writes over the Ch channel.
type dummyConn struct {
	Ch       chan []byte
	IsClosed bool
	mu       sync.Mutex
}

// Larger than handlers.BufSize for testing purposes
const dummyBufSize = handlers.BufSize * 2

// NewDummyConn returns an initialized dummyConn
func NewDummyConn() *dummyConn {
	return &dummyConn{
		make(chan []byte),
		false,
		sync.Mutex{}}
}

// Read reads from Ch and stores the read bytes in b.  It returns the number
// of bytes read and any errors.
func (d *dummyConn) Read(b []byte) (n int, err error) {
	if len(b) > dummyBufSize {
		return 0, errors.New(fmt.Sprintf(
			"Message too large to read (%d)", len(b)))
	}
	if d.IsClosed {
		return 0, errors.New("Connection is closed")
	}
	out := <-d.Ch
	if out == nil {
		return 0, errors.New("Connection closed")
	}
	n = copy(b, out)
	return n, nil
}

// Write writes the bytes from b into Ch.  It returns the number of bytes
// written and any errors.
func (d *dummyConn) Write(b []byte) (n int, err error) {
	// We don't want to write to a closed channel
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(b) > dummyBufSize {
		return 0, errors.New(fmt.Sprintf(
			"Message too large to write (%d)", len(b)))
	}
	if d.IsClosed {
		return 0, errors.New("Connection is closed")
	}
	b = b[:]
	d.Ch <- b
	return len(b), nil
}

// Close marks the dummyConn as closed and closes Ch
func (d *dummyConn) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.IsClosed {
		close(d.Ch)
		d.IsClosed = true
	}
	return nil
}

// dummyAddr implements net.Addr.  It is meant for testing purposes.
type dummyAddr struct {
	network string
	str     string
}

// Network returns the network string
func (d *dummyAddr) Network() string {
	return d.network
}

// Network returns the str string
func (d *dummyAddr) String() string {
	return d.str
}

// LocalAddr returns the "local" dummyAddr
func (d *dummyConn) LocalAddr() net.Addr {
	return &dummyAddr{"dummy", "local"}
}

// LocalAddr returns the "remote" dummyAddr
func (d *dummyConn) RemoteAddr() net.Addr {
	return &dummyAddr{"dummy", "remote"}
}

// SetDeadline is not implemented for dummyConn
func (d *dummyConn) SetDeadline(t time.Time) error {
	return nil
}

// SetDeadline is not implemented for dummyConn
func (d *dummyConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetDeadline is not implemented for dummyConn
func (d *dummyConn) SetWriteDeadline(t time.Time) error {
	return nil
}
