package dummyconn

import (
	"errors"
	"net"
	"sync"
	"time"
)

var ConnClosedErrRead = errors.New("Read error: connection closed")
var ConnClosedErrWrite = errors.New("Write error: connection closed")

// dummyConn implements net.Conn.  It is meant to use as a mock connection
// for unit tests.  It communicates reads and writes over the Ch channel.
type dummyConn struct {
	Ch       chan []byte
	isClosed bool
	mu       sync.Mutex
}

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
	out := <-d.Ch
	if out == nil {
		return 0, ConnClosedErrRead
	}
	n = copy(b, out)
	return n, nil
}

// Write writes the bytes from b into Ch.  It returns the number of bytes
// written and any errors.
func (d *dummyConn) Write(b []byte) (n int, err error) {
	defer func() {
		// Return error if the "connection" (i.e., the channel) is closed
		if e := recover(); e != nil {
			err = ConnClosedErrWrite
		}
	}()
	b = b[:]
	d.Ch <- b
	return len(b), nil
}

// Close marks the dummyConn as closed and closes Ch.  Close() can be called
// multiple times.
func (d *dummyConn) Close() error {
	defer func() {
		// Channel already closed, that's OK
		recover()
	}()
	close(d.Ch)
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
