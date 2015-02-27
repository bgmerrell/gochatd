package http

import (
	"io/ioutil"
	"net/http"

	"github.com/bgmerrell/gochatd/chat"
)

const (
	nameParam = "name"
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

func get(w http.ResponseWriter, r *http.Request) (hndlErr *HandlerError) {
	return handlerErrorFromCode(http.StatusNotImplemented)
}

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

func Handle(w http.ResponseWriter, r *http.Request, cm *chat.ChatManager, maxBodySize int, maxNameSize int) (hndlErr *HandlerError) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBodySize))
	if r.Method == "GET" {
		hndlErr = get(w, r)
	} else if r.Method == "POST" {
		hndlErr = post(w, r, cm, maxNameSize)
	} else {
		// HTTP/1.1 spec says we must indicate which methods we allow
		w.Header().Set("Allow", "GET, POST")
		hndlErr = handlerErrorFromCode(http.StatusMethodNotAllowed)
	}
	return hndlErr
}
