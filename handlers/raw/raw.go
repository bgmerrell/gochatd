package raw

import (
	"log"
	"net"
	"strings"

	"github.com/bgmerrell/gochatd/handlers"
)

func Handle(conn net.Conn) {
	_, err := conn.Write([]byte("What's your name?: "))
	if err != nil {
		log.Println("Error requesting name:", err.Error())
		return
	}
	buf := make([]byte, handlers.BufSize)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println(err)
		conn.Close()
	}
	log.Println("buf", buf)
	name := strings.TrimSpace(string(buf[:n]))
	log.Printf("name: \"%s\"", name)
	if name == "foo" {
		log.Println("no foos allowed")
		_, _ = conn.Write([]byte("No foos allowed!\n"))
		conn.Close()
		return
	}
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			conn.Close()
			return
		}
		log.Println(string(buf[:n]))
	}
}
