package chat

import (
	"bufio"
	"bytes"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bgmerrell/gochatd/dummyconn"
)

const bufSize = 512
const historySize = 8
const testTime = "02-Jan-06 15:04"

func init() {
	Timestamp = func() string { return testTime }
}

func TestJoin(t *testing.T) {
	// rwBuf := bufio.NewReadWriter([]byte{}, []byte{})
	cm := NewChatManager(nil, historySize)
	dc := dummyconn.NewDummyConn()
	err := cm.Join("testuser", dc)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	buf := make([]byte, bufSize)
	n, err := dc.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg := buf[:n]
	expected := []byte(testTime + " * testuser has joined\n")
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want %s.", rMsg, expected)
	}
}

func TestJoinDuplicateUser(t *testing.T) {
	cm := NewChatManager(nil, historySize)
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
	cm := NewChatManager(nil, historySize)
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
	logBuf := &bytes.Buffer{}
	expectedChatLog := []byte{}
	writer := bufio.NewWriter(logBuf)
	cm := NewChatManager(writer, historySize)
	dc1 := dummyconn.NewDummyConn()
	dc2 := dummyconn.NewDummyConn()
	readBuf := make([]byte, bufSize)

	err := cm.Join("testuser1", dc1)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	n, err := dc1.Read(readBuf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg := readBuf[:n]
	expected := []byte(testTime + " * testuser1 has joined\n")
	expectedChatLog = append(expectedChatLog, expected...)
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want: %s.", rMsg, expected)
	}

	err = cm.Join("testuser2", dc2)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	// Both joined users will see the testuser2 join message.
	n, err = dc1.Read(readBuf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg = readBuf[:n]
	expected = []byte(testTime + " * testuser2 has joined\n")
	expectedChatLog = append(expectedChatLog, expected...)
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want: %s.", rMsg, expected)
	}
	n, err = dc2.Read(readBuf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg = readBuf[:n]
	expected = []byte(testTime + " * testuser2 has joined\n")
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want: %s.", rMsg, expected)
	}

	cm.Broadcast("testuser", []byte("test message"))

	n, err = dc1.Read(readBuf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg = readBuf[:n]
	expected = []byte(testTime + " <testuser> test message\n")
	expectedChatLog = append(expectedChatLog, expected...)
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want: %s.", rMsg, expected)
	}

	n, err = dc2.Read(readBuf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg = readBuf[:n]
	expected = []byte(testTime + " <testuser> test message\n")
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want %s.", rMsg, expected)
	}

	writer.Flush()
	actualChatLog := logBuf.String()
	if actualChatLog != string(expectedChatLog) {
		t.Fatalf("Unexpected chat log contents: %s, want: %s.",
			actualChatLog, string(expectedChatLog))
	}
}

func TestDefaultTimestamp(t *testing.T) {
	timeString := _timestamp()
	_, err := time.Parse(timestampLayout, timeString)
	if err != nil {
		t.Fatal("Failed to parse time string: %s", timeString)
	}
}

type FailWriter struct{}

func (f *FailWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("FailWriter always fails")
}

// TestLogWriteFail makes sure that a message is still broadcast even if the
// message fails to be written to the chat log.
func TestLogWriteFail(t *testing.T) {
	cm := NewChatManager(&FailWriter{}, historySize)
	dc := dummyconn.NewDummyConn()
	err := cm.Join("testuser", dc)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	buf := make([]byte, bufSize)
	n, err := dc.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	rMsg := buf[:n]
	expected := []byte(testTime + " * testuser has joined\n")
	if !bytes.Equal(rMsg, expected) {
		t.Fatalf("Unexpected read: %s, want %s.", rMsg, expected)
	}
}

func TestHistoryInsert(t *testing.T) {
	h := newHistory(historySize)
	msg := []byte("0")
	h.insert(msg)
	if h.message.Value != nil {
		t.Error("Expected message position to be nil")
	}
	prev := h.message.Prev()
	if !bytes.Equal(prev.Value.([]byte), msg) {
		t.Errorf("message = %s, want: %s", prev.Value.([]byte), msg)
	}
}

func TestHistoryInsertFull(t *testing.T) {
	h := newHistory(historySize)
	for i := 0; i < historySize; i++ {
		h.insert([]byte(strconv.Itoa(i)))
	}
	expected := []byte("0")
	if !bytes.Equal(h.message.Value.([]byte), expected) {
		t.Errorf("message = %s, want: %s", h.message.Value.([]byte), expected)
	}
	expected = []byte("7")
	prev := h.message.Prev()
	if !bytes.Equal(prev.Value.([]byte), expected) {
		t.Errorf("message = %s, want: %s", prev.Value.([]byte), expected)
	}
}

func TestHistoryMessages(t *testing.T) {
	h := newHistory(historySize)
	for i := 0; i < historySize; i++ {
		h.insert([]byte(strconv.Itoa(i)))
	}
	expected := []byte("01234567")
	messages := h.messages(historySize + 1)
	if !bytes.Equal(messages, expected) {
		t.Errorf("message = %s, want: %s", messages, expected)
	}
}

func TestHistoryMessagesFullPlusOne(t *testing.T) {
	h := newHistory(historySize)
	for i := 0; i < historySize+1; i++ {
		h.insert([]byte(strconv.Itoa(i)))
	}
	expected := []byte("12345678")
	messages := h.messages(historySize)
	if !bytes.Equal(messages, expected) {
		t.Errorf("message = %s, want: %s", messages, expected)
	}
}

func TestHistoryMessagesNotFull(t *testing.T) {
	h := newHistory(historySize)
	for i := 0; i < historySize-1; i++ {
		h.insert([]byte(strconv.Itoa(i)))
	}
	expected := []byte("0123456")
	messages := h.messages(historySize)
	if !bytes.Equal(messages, expected) {
		t.Errorf("message = %s, want: %s", messages, expected)
	}
}

func TestHistoryMessagesEmpty(t *testing.T) {
	h := newHistory(historySize)
	expected := []byte("")
	messages := h.messages(historySize)
	if !bytes.Equal(messages, expected) {
		t.Errorf("message = %s, want: %s", messages, expected)
	}
}

func TestHistory(t *testing.T) {
	cm := NewChatManager(nil, historySize)
	for i := 0; i < historySize; i++ {
		cm.history.insert([]byte(strconv.Itoa(i)))
	}
	expected := []byte("01234567")
	messages := cm.History(historySize)
	if !bytes.Equal(messages, expected) {
		t.Errorf("message = %s, want: %s", messages, expected)
	}
}
