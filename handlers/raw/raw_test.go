package raw

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/bgmerrell/gochatd/chat"
	"github.com/bgmerrell/gochatd/dummyconn"
)

var bufSize int = 512

const historySize = 8

var maxNameSize int = 32

var wg sync.WaitGroup

func TestHandle(t *testing.T) {
	cm := chat.NewChatManager(nil, historySize)
	dc := dummyconn.NewDummyConn()
	rh := NewRawHandler(bufSize, maxNameSize)
	wg.Add(1)
	go func() {
		defer wg.Done()
		rh.Handle(cm, dc)
	}()
	buf := make([]byte, bufSize)
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
	cm := chat.NewChatManager(nil, historySize)
	dc := dummyconn.NewDummyConn()
	rh := NewRawHandler(bufSize, maxNameSize)
	wg.Add(1)
	go func() {
		defer wg.Done()
		rh.Handle(cm, dc)
	}()
	buf := make([]byte, bufSize)
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
	cm := chat.NewChatManager(nil, historySize)
	dc := dummyconn.NewDummyConn()
	rh := NewRawHandler(bufSize, maxNameSize)
	wg.Add(1)
	go func() {
		defer wg.Done()
		rh.Handle(cm, dc)
	}()
	buf := make([]byte, bufSize)
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
	rh := NewRawHandler(bufSize, maxNameSize)
	dc.Close()
	_, err := rh.getName(dc)
	if !strings.HasPrefix(err.Error(), "Error requesting name") {
		t.Error("Expected error requesting name")
	}
}

func TestHandleErrReadingName(t *testing.T) {
	cm := chat.NewChatManager(nil, historySize)
	dc := dummyconn.NewDummyConn()
	rh := NewRawHandler(bufSize, maxNameSize)
	wg.Add(1)
	go func() {
		defer wg.Done()
		rh.Handle(cm, dc)
	}()
	buf := make([]byte, bufSize)
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

func TestHandleDuplicateUser(t *testing.T) {
	cm := chat.NewChatManager(nil, historySize)
	rh := NewRawHandler(bufSize, maxNameSize)
	dc1 := dummyconn.NewDummyConn()
	dc2 := dummyconn.NewDummyConn()
	wg.Add(1)
	go func() {
		defer wg.Done()
		rh.Handle(cm, dc1)
	}()
	go func() {
		defer wg.Done()
		rh.Handle(cm, dc2)
	}()
	buf := make([]byte, bufSize)

	// "testuser" logging in on dc1
	n, err := dc1.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg := buf[:n]
	expected := []byte(namePrompt)
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want: %s.", rMsg, expected)
	}
	wMsg := []byte("testuser\r\n")
	n, err = dc1.Write(wMsg)
	expectedN := len(wMsg)
	if n != expectedN {
		t.Error("n: %d, want: %d.", n, expectedN)
	}

	// "testuser" logging in on dc2
	n, err = dc2.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg = buf[:n]
	expected = []byte(namePrompt)
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want: %s.", rMsg, expected)
	}
	wMsg = []byte("testuser\r\n")
	n, err = dc2.Write(wMsg)
	expectedN = len(wMsg)
	if n != expectedN {
		t.Error("n: %d, want: %d.", n, expectedN)
	}

	// mock client disconnect of dc1; no need to disconnect dc2 due to
	// duplicate user.
	err = dc1.Close()

	if err != nil {
		t.Fatal("Error closing:", err.Error())
	}
	wg.Wait()
}
