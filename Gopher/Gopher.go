package gopher

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	listing "../Libs/Listing"
)

//Server is a structor to contain the config information for the server.
type Server struct {
	Log  string
	Host string
	Port string
	Root string
}

//Listen inits server
func (server *Server) Listen(customHandlerFunc func(*string, *string, *net.TCPConn)) error {
	listen, _error := net.Listen("tcp", server.Host+":"+server.Port)
	if _error != nil {
		log("ERROR", _error.Error()+"\n", server.Log)
		return _error
	}

	log("INFO", "Listening on gopher://"+server.Host+":"+server.Port+"\n", server.Log)

	for {
		connection, _error := listen.Accept()
		if _error != nil {
			log("ERROR", _error.Error()+"\n", server.Log)
			return _error
		}

		go server.serve(connection.(*net.TCPConn), customHandlerFunc)
	}
}

func (server *Server) serve(connection *net.TCPConn, customHandlerFunc func(*string, *string, *net.TCPConn)) {
	defer connection.Close()
	path, param, _error := server.parseURL(connection)
	if _error != nil {
		log("ERROR", _error.Error()+"\n", server.Log)
		return
	}
	server.handler(&path, &param, connection, false, customHandlerFunc)
}

func (server *Server) handler(path *string, parameter *string, connection *net.TCPConn, customHandler bool,
	customHandlerFunc func(*string, *string, *net.TCPConn)) {
	if customHandler {
		customHandlerFunc(path, parameter, connection)
	} else {
		server.defaultHandler(path, connection)
	}
}

func (server *Server) defaultHandler(path *string, connection *net.TCPConn) {
	log("INFO", "Serve: "+*path+"\n", server.Log)
	if *path == "/" {
		server.sendFile("/index", connection)
	} else {
		filename := server.filename(*path)
		fileindex, _error := os.Stat(filename)
		if _error != nil {
			log("ERROR", _error.Error()+"\n", server.Log)
			return
		}
		if fileindex.IsDir() {
			var list listing.Listing

			filepath.Walk(filename, func(path string, info os.FileInfo, _error error) error {
				if info.IsDir() {
					return list.AppendDir(info.Name(), path, server.Root, server.Host, server.Port)
				}

				list.AppendFile(info.Name(), path, server.Root, server.Host, server.Port)
				return nil
			})

			fmt.Fprint(connection, list)
			return
		}
		server.sendFile(*path, connection)
	}
}

func (server *Server) sendFile(path string, connection *net.TCPConn) {
	filename := server.filename(path)
	file, _error := os.Open(filename)
	if _error != nil {
		log("ERROR", _error.Error()+"\n", server.Log)
		return
	}

	defer file.Close()

	connection.ReadFrom(file)
}

//Helper functions
func log(infoType string, info string, log string) {
	currentTime := "[" + time.Now().Format("2006-01-02 15:04:05") + "] "
	fmt.Println(currentTime + "[" + infoType + "] " + info)
	file, _error := os.OpenFile(log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if _error != nil {
		fmt.Println(currentTime + "[ERROR] " + _error.Error() + "\n")
		return
	}
	if _, _error := file.Write([]byte(currentTime + "[" + infoType + "] " + info)); _error != nil {
		fmt.Println(currentTime + "[ERROR] " + _error.Error() + "\n")
		return
	}
	if _error := file.Close(); _error != nil {
		fmt.Println(currentTime + "[ERROR] " + _error.Error() + "\n")
		return
	}
	return
}

func (server *Server) filename(path string) string {
	cPath := filepath.Clean("/" + path)
	return server.Root + cPath
}

func (server *Server) parseURL(reader io.Reader) (string, string, error) {
	buf := make([]byte, 512)
	length, _error := reader.Read(buf)
	if _error != nil {
		log("ERROR", _error.Error()+"\n", server.Log)
		return "", "", _error
	}
	parsed := strings.Split(string(buf[:length]), "	")
	if len(parsed) > 1 {
		log("INFO", "Parsed url: "+parsed[0]+" Parameter: "+parsed[1]+"\n", server.Log)
		return parsed[0], parsed[1], _error
	}
	return strings.TrimSpace(parsed[0]), "", _error
}
