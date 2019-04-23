package main

import gopher "./Gopher"

func main() {
	server := gopher.Server{"./log/log.txt", "192.168.1.114", "70", "7070", "./Root"}
	server.Listen()
}
