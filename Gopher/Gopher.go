package gopher

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
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
	Log     string
	Host    string
	Port    string
	PortTLS string
	Root    string
}

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

//ListenTLS inits TLS server
func (server *Server) ListenTLS() error {
	certificate, _error := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if _error != nil {
		log("ERROR", _error.Error()+"\n", server.Log)
		return _error
	}
	config := tls.Config{Certificates: []tls.Certificate{certificate}}
	config.Rand = rand.Reader
	listenTLS, _error := tls.Listen("tcp", server.Host+":"+server.PortTLS, &config)
	if _error != nil {
		log("ERROR", _error.Error()+"\n", server.Log)
		return _error
	}
	log("INFO", "Listening on gophers://"+server.Host+":"+server.PortTLS+"\n", server.Log)
	for {
		connectionTLS, _error := listenTLS.Accept()
		if _error != nil {
			log("ERROR", _error.Error()+"\n", server.Log)
			return _error
		}
		log("INFO", "Connection accepted"+"\n", server.Log)

		defer connectionTLS.Close()

		TLSconnection, ok := connectionTLS.(*tls.Conn)
		if ok {
			state := TLSconnection.ConnectionState()
			for _, v := range state.PeerCertificates {
				_byte, _ := x509.MarshalPKIXPublicKey(v.PublicKey)
				log("INFO", string(_byte), server.Log)
			}
		}
		go server.serveTLS(connectionTLS.(*tls.Conn))
	}
}

//Listen inits server
func (server *Server) Listen() error {
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

		go server.serve(connection.(*net.TCPConn))
	}
}

func (server *Server) serveTLS(connection *tls.Conn) {
	defer connection.Close()
	path, param, _error := server.parseURL(connection)
	if _error != nil {
		log("ERROR", _error.Error()+"\n", server.Log)
		return
	}
	log("INFO", "ServeTLS: "+path+"\n", server.Log)
	server.handlerTLS(&path, &param, connection)
}

func (server *Server) serve(connection *net.TCPConn) {
	defer connection.Close()
	path, param, _error := server.parseURL(connection)
	if _error != nil {
		log("ERROR", _error.Error()+"\n", server.Log)
		return
	}
	server.handler(&path, &param, connection)
}

func (server *Server) handlerTLS(path *string, param *string, connection *tls.Conn) {
	log("INFO", "handlerTLS: "+*path+"\n", server.Log)
	if *path == "/" {
		server.sendFileTLS("/index", connection)
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
		server.sendFileTLS(*path, connection)
	}
}

func (server *Server) handler(path *string, param *string, connection *net.TCPConn) {
	log("INFO", "Serve: "+*path+"\n", server.Log)
	if *path == "/" {
		server.sendFile("/index", connection)
	} else {
		filename := server.filename(*path)
		fileindex, _error := os.Stat(filename)
		if _error != nil {
			log("ERRORf", _error.Error()+"\n", server.Log)
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

func (server *Server) sendFileTLS(path string, connection *tls.Conn) {
	filename := server.filename(path)
	file, _error := os.Open(filename)
	if _error != nil {
		log("ERROR", _error.Error()+"\n", server.Log)
		return
	}

	defer file.Close()
	fileInfo, _error := file.Stat()
	if _error != nil {
		log("ERROR", _error.Error()+"\n", server.Log)
		return
	}
	size := fileInfo.Size()
	bytes := make([]byte, size)
	buffer := bufio.NewReader(file)
	_, _error = buffer.Read(bytes)
	if _error != nil {
		log("ERROR", _error.Error()+"\n", server.Log)
		return
	}
	connection.Write(bytes)
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
