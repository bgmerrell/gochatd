package dummyconn

import (
	"bytes"
	"testing"
	"time"

	"github.com/bgmerrell/gochatd/handlers"
)

func TestWriteRead(t *testing.T) {
	dc := NewDummyConn()
	wMsg := []byte("foo")
	go func() {
		n, err := dc.Write(wMsg)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if n != len(wMsg) {
			t.Errorf("n = %d, want %d", n, len(wMsg))
		}
	}()
	buf := make([]byte, handlers.BufSize)
	n, err := dc.Read(buf)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if n != len(wMsg) {
		t.Errorf("n = %d, want %d", n, len(wMsg))
	}
	rMsg := buf[:n]
	if !bytes.Equal(rMsg, wMsg) {
		t.Errorf("read message: %s, want %s", n, len(wMsg))
	}
}

func TestWriteDisconnected(t *testing.T) {
	dc := NewDummyConn()
	dc.Close()
	wMsg := []byte("foo")
	_, err := dc.Write(wMsg)
	if err != ConnClosedErrWrite {
		t.Error("Expected closed connection error")
	}
}

func TestReadDisconnected(t *testing.T) {
	dc := NewDummyConn()
	dc.Close()
	rMsg := []byte("foo")
	_, err := dc.Read(rMsg)
	if err != ConnClosedErrRead {
		t.Error("Expected closed connection error")
	}
}

func TestLocalAddr(t *testing.T) {
	dc := NewDummyConn()
	la := dc.LocalAddr()
	expectedNetwork := "dummy"
	expectedString := "local"
	if la.Network() != expectedNetwork {
		t.Errorf("Network() = %s, want: %s", la.Network(), expectedNetwork)
	}
	if la.String() != expectedString {
		t.Errorf("String() = %s, want: %s", la.String(), expectedString)
	}
}

func TestRemoteAddr(t *testing.T) {
	dc := NewDummyConn()
	ra := dc.RemoteAddr()
	expectedNetwork := "dummy"
	expectedString := "remote"
	if ra.Network() != expectedNetwork {
		t.Errorf("Network() = %s, want: %s", ra.Network(), expectedNetwork)
	}
	if ra.String() != expectedString {
		t.Errorf("String() = %s, want: %s", ra.String(), expectedString)
	}
}

func TestSetDeadline(t *testing.T) {
	dc := NewDummyConn()
	err := dc.SetDeadline(time.Now())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestSetReadDeadline(t *testing.T) {
	dc := NewDummyConn()
	err := dc.SetReadDeadline(time.Now())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestSetWriteDeadline(t *testing.T) {
	dc := NewDummyConn()
	err := dc.SetWriteDeadline(time.Now())
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}
