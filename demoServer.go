package main

import (
	"net"

	gopher "./Gopher"
)

func main() {
	server := gopher.Server{"./log/log.txt", "192.168.1.147", "70", "./Root"}
	server.Listen(func(*string, *string, *net.TCPConn) {})
}
