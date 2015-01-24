package raw

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/bgmerrell/gochatd/dummyconn"
	"github.com/bgmerrell/gochatd/handlers"
)

var wg sync.WaitGroup

func TestHandle(t *testing.T) {
	dc := dummyconn.NewDummyConn()
	wg.Add(1)
	go func() {
		defer wg.Done()
		Handle(dc)
	}()
	buf := make([]byte, handlers.BufSize)
	n, err := dc.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg := buf[:n]
	expected := []byte(namePrompt)
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want %s.", rMsg, expected)
	}

	wMsg := []byte("testuser\r\n")
	n, err = dc.Write(wMsg)
	expectedN := len(wMsg)
	if n != expectedN {
		t.Error("n: %d, want: %d.", n, expectedN)
	}

	wMsg = []byte("A test message\r\n")
	n, err = dc.Write(wMsg)
	expectedN = len(wMsg)
	if n != expectedN {
		t.Error("n: %d, want: %d.", n, expectedN)
	}

	// mock client disconnecting
	err = dc.Close()

	if err != nil {
		t.Fatal("Error closing:", err.Error())
	}
	wg.Wait()
}

func TestHandleEmptyName(t *testing.T) {
	dc := dummyconn.NewDummyConn()
	wg.Add(1)
	go func() {
		defer wg.Done()
		Handle(dc)
	}()
	buf := make([]byte, handlers.BufSize)
	n, err := dc.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg := buf[:n]
	expected := []byte(namePrompt)
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want %s.", rMsg, expected)
	}

	// Empty name, i.e., user pressed "return" with no other input
	wMsg := []byte("\r\n")
	n, err = dc.Write(wMsg)
	expectedN := len(wMsg)
	if n != expectedN {
		t.Error("n: %d, want: %d.", n, expectedN)
	}

	n, err = dc.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg = buf[:n]
	if !bytes.HasPrefix(rMsg, []byte("Disconnecting")) {
		t.Errorf("Read: %s, expected \"Disconnecting\" prefix", rMsg)
	}

	wg.Wait()
}

func TestHandleLongName(t *testing.T) {
	dc := dummyconn.NewDummyConn()
	wg.Add(1)
	go func() {
		defer wg.Done()
		Handle(dc)
	}()
	buf := make([]byte, handlers.BufSize)
	n, err := dc.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg := buf[:n]
	expected := []byte(namePrompt)
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want %s.", rMsg, expected)
	}

	wMsg := []byte("thisnameisjusttoolongtobereasonable\r\n")
	n, err = dc.Write(wMsg)
	expectedN := len(wMsg)
	if n != expectedN {
		t.Error("n: %d, want: %d.", n, expectedN)
	}

	n, err = dc.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg = buf[:n]
	if !bytes.HasPrefix(rMsg, []byte("Disconnecting")) {
		t.Errorf("Read: %s, expected \"Disconnecting\" prefix", rMsg)
	}

	wg.Wait()
}

func TestGetNameErrRequesting(t *testing.T) {
	dc := dummyconn.NewDummyConn()
	dc.Close()
	_, err := getName(dc)
	if !strings.HasPrefix(err.Error(), "Error requesting name") {
		t.Error("Expected error requesting name")
	}
}

func TestHandleErrReadingName(t *testing.T) {
	dc := dummyconn.NewDummyConn()
	wg.Add(1)
	go func() {
		defer wg.Done()
		Handle(dc)
	}()
	buf := make([]byte, handlers.BufSize)
	n, err := dc.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg := buf[:n]
	expected := []byte(namePrompt)
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want %s.", rMsg, expected)
	}

	dc.Close()

	wg.Wait()
}
