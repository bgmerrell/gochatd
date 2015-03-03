package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bgmerrell/gochatd/chat"
)

const (
	historySize int = 4
	testTime        = "02-Jan-06 15:04"
)

func TestGet(t *testing.T) {
	chat.Timestamp = func() string { return testTime }
	cm := chat.NewChatManager(nil, historySize)
	cm.Broadcast("user1", []byte("1"))
	cm.Broadcast("user2", []byte("2"))
	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	hErr := get(w, req, cm, historySize)
	if hErr != nil {
		t.Fatal("Unexpected error: ", err)
	}
	if w.Code != http.StatusOK {
		t.Errorf("Response code = %d, want: %d", w.Code, http.StatusOK)
	}
	expected := testTime + " <user1> 1\n" + testTime + " <user2> 2\n"
	if w.Body.String() != expected {
		t.Errorf("Response body = %s, want: %s", w.Body.String(), expected)
	}
}

// TODO: Add more tests
