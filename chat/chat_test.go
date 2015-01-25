package chat

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bgmerrell/gochatd/dummyconn"
	"github.com/bgmerrell/gochatd/handlers"
)

func TestJoin(t *testing.T) {
	cm := NewChatManager()
	dc := dummyconn.NewDummyConn()
	err := cm.Join("testuser", dc)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	buf := make([]byte, handlers.BufSize)
	n, err := dc.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg := buf[:n]
	expected := []byte("* testuser has joined\n")
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want %s.", rMsg, expected)
	}
}

func TestJoinDuplicateUser(t *testing.T) {
	cm := NewChatManager()
	dc := dummyconn.NewDummyConn()
	err := cm.Join("testuser", dc)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	err = cm.Join("testuser", dc)
	if err == nil || !strings.HasSuffix(err.Error(), "already connected") {
		if err == nil {
			t.Error("Expected error due to duplicate user")
		} else {
			t.Error("Expected error due to duplicate user, got:", err.Error())
		}
	}
}

func TestQuit(t *testing.T) {
	cm := NewChatManager()
	dc := dummyconn.NewDummyConn()
	err := cm.Join("testuser", dc)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	cm.Quit("testuser")

	err = cm.Join("testuser", dc)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestBroadcast(t *testing.T) {
	cm := NewChatManager()
	dc1 := dummyconn.NewDummyConn()
	dc2 := dummyconn.NewDummyConn()
	buf := make([]byte, handlers.BufSize)

	err := cm.Join("testuser1", dc1)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	n, err := dc1.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg := buf[:n]
	expected := []byte("* testuser1 has joined\n")
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want: %s.", rMsg, expected)
	}

	err = cm.Join("testuser2", dc2)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	// Both joined users will see the testuser2 join message.
	n, err = dc1.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg = buf[:n]
	expected = []byte("* testuser2 has joined\n")
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want: %s.", rMsg, expected)
	}
	n, err = dc2.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg = buf[:n]
	expected = []byte("* testuser2 has joined\n")
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want: %s.", rMsg, expected)
	}

	cm.Broadcast("test message\n")

	n, err = dc1.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg = buf[:n]
	expected = []byte("test message\n")
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want: %s.", rMsg, expected)
	}

	n, err = dc2.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg = buf[:n]
	expected = []byte("test message\n")
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want %s.", rMsg, expected)
	}
}
