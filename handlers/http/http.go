package http

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/bgmerrell/gochatd/chat"
)

const (
	nameParam        = "name"
	linesParam       = "lines"
	minLinesParamVal = 1
)

type HandlerError struct {
	Code int
	Msg  string
}

func handlerErrorFromCode(code int) *HandlerError {
	return &HandlerError{
		code,
		http.StatusText(code),
	}
}

// get reads from the chat.  The HTTP requests's "lines" parameter is used to
// specify the number of lines to read from the chat.
func get(w http.ResponseWriter, r *http.Request, cm *chat.ChatManager, maxHistoryLines int) (hndlErr *HandlerError) {
	linesParamVal := r.URL.Query().Get(linesParam)
	numLines, err := strconv.Atoi(linesParamVal)
	// If a lines parameter was invalid or missing, just ask for all of
	// the history lines.
	if err != nil || numLines < minLinesParamVal {
		numLines = maxHistoryLines
	}
	_, err = w.Write(cm.History(numLines))
	if err != nil {
		return &HandlerError{http.StatusInternalServerError, err.Error()}
	}
	return hndlErr
}

// post posts a message (the HTTP body) to the chat
func post(w http.ResponseWriter, r *http.Request, cm *chat.ChatManager, maxNameSize int) (hndlErr *HandlerError) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return &HandlerError{http.StatusInternalServerError, err.Error()}
	}
	name := r.FormValue(nameParam)
	if name == "" {
		return &HandlerError{http.StatusBadRequest, "missing name"}
	} else if len(name) > maxNameSize {
		return &HandlerError{http.StatusBadRequest, "name too long"}
	}
	cm.Broadcast(name, body)
	return hndlErr
}

// Handle supports HTTP writing (via POST) and reading (via GET) to the chat.
func Handle(w http.ResponseWriter, r *http.Request, cm *chat.ChatManager, maxBodySize int, maxNameSize int, maxHistoryLines int) (hndlErr *HandlerError) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBodySize))
	if r.Method == "GET" {
		hndlErr = get(w, r, cm, maxHistoryLines)
	} else if r.Method == "POST" {
		hndlErr = post(w, r, cm, maxNameSize)
	} else {
		// HTTP/1.1 spec says we must indicate which methods we allow
		w.Header().Set("Allow", "GET, POST")
		hndlErr = handlerErrorFromCode(http.StatusMethodNotAllowed)
	}
	return hndlErr
}
