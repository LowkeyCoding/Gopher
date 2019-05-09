package main

import (
	"crypto/tls"

	gophers "./Gophers"
)

func main() {
	server := gophers.ServerTLS{"./log/log.txt", "192.168.1.114", "7070", "./Root", "./certs"}
	server.Listen(func(*string, *string, *tls.Conn) {})
}
