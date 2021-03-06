package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/bgmerrell/gochatd/chat"
	httphandler "github.com/bgmerrell/gochatd/handlers/http"
	"github.com/bgmerrell/gochatd/handlers/raw"
)

var confPath string

func init() {
	flag.StringVar(&confPath, "conf-path", "gochatd-conf.json", "Configuration file path")
}

type config struct {
	LogPath         string `json:"log_path"`
	Addr            string `json:"address"`
	MaxNameLen      int    `json:"max_name_length"`
	MsgBufSize      int    `json:"msg_buffer_size"`
	MaxHistoryLines int    `json:"max_history_lines"`
}

func main() {
	flag.Parse()
	cfgRaw, err := ioutil.ReadFile(confPath)
	if err != nil {
		log.Fatalf("Failed to read log file (%s): %s", confPath, err)
	}
	// Just use a JSON config file to avoid 3rd party dependencies (for
	// something like ini or toml)
	cfg := config{}
	err = json.Unmarshal(cfgRaw, &cfg)
	if err != nil {
		log.Fatalf("Failed to parse log file (%s): %s", confPath, err)
	}
	chatLogFile, err := os.OpenFile(cfg.LogPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Failed to open chat log: %s", err)
	}
	defer chatLogFile.Close()
	cm := chat.NewChatManager(chatLogFile, cfg.MaxHistoryLines)

	http.HandleFunc("/chat",
		func(w http.ResponseWriter, r *http.Request) {
			hndlErr := httphandler.Handle(w, r, cm, cfg.MsgBufSize, cfg.MaxNameLen, cfg.MaxHistoryLines)
			if hndlErr != nil {
				log.Print(hndlErr.Msg)
				http.Error(w, hndlErr.Msg, hndlErr.Code)
			}
		})
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		rh := raw.NewRawHandler(cfg.MsgBufSize, cfg.MaxNameLen)
		conn, err := ln.Accept()
		if err != nil {

		}
		go rh.Handle(cm, conn)
	}
}
