package main

import "./Gopher"

func main() {
	server := Gopher.Server{"./log/log.txt", "192.168.1.114", "70", "7070", "./Root"}
	server.Listen()
}
