package dummyconn

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/bgmerrell/gochatd/handlers"
)

type dummyConn struct {
	read     []byte
	written  []byte
	isClosed bool
}

// Larger than handlers.BufSize for testing purposes
const dummyBufSize = handlers.BufSize * 2

func NewDummyConn() *dummyConn {
	return &dummyConn{
		make([]byte, dummyBufSize),
		make([]byte, dummyBufSize),
		false}
}

func (d *dummyConn) Read(b []byte) (n int, err error) {
	if len(b) > dummyBufSize {
		return 0, errors.New(fmt.Sprintf(
			"Message too large to read (%d)", len(b)))
	}
	d.read = d.written[:]
	return len(d.read), nil
}

func (d *dummyConn) Write(b []byte) (n int, err error) {
	if len(b) > dummyBufSize {
		return 0, errors.New(fmt.Sprintf(
			"Message too large to write (%d)", len(b)))
	}
	d.written = b[:]
	return len(d.written), nil
}

func (d *dummyConn) Close() error {
	d.isClosed = true
	return nil
}

type dummyAddr struct {
	network string
	str     string
}

func (d *dummyAddr) Network() string {
	return d.network
}

func (d *dummyAddr) String() string {
	return d.str
}

func (d *dummyConn) LocalAddr() net.Addr {
	return &dummyAddr{"dummy", "local"}
}

func (d *dummyConn) RemoteAddr() net.Addr {
	return &dummyAddr{"dummy", "remote"}
}

func (d *dummyConn) SetDeadline(t time.Time) error {
	return nil
}

func (d *dummyConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (d *dummyConn) SetWriteDeadline(t time.Time) error {
	return nil
}
