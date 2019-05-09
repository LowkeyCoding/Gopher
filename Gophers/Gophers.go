package gophers

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	listing "../Libs/Listing"
)

type ServerTLS struct {
	Log          string
	Host         string
	Port         string
	Root         string
	Certificates string
}

//ListenTLS inits TLS server
func (server *ServerTLS) Listen(customHandlerFunc func(*string, *string, *tls.Conn)) error {
	certificate, _error := tls.LoadX509KeyPair(server.Certificates+"/server.pem", server.Certificates+"/server.key")
	if _error != nil {
		log("ERROR", _error.Error()+"\n", server.Log)
		return _error
	}
	config := tls.Config{Certificates: []tls.Certificate{certificate}}
	config.Rand = rand.Reader
	listenTLS, _error := tls.Listen("tcp", server.Host+":"+server.Port, &config)
	if _error != nil {
		log("ERROR", _error.Error()+"\n", server.Log)
		return _error
	}
	log("INFO", "Listening on gophers://"+server.Host+":"+server.Port+"\n", server.Log)
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
		go server.serve(connectionTLS.(*tls.Conn), customHandlerFunc)
	}
}

func (server *ServerTLS) serve(connection *tls.Conn, customHandlerFunc func(*string, *string, *tls.Conn)) {
	defer connection.Close()
	path, param, _error := server.parseURL(connection)
	if _error != nil {
		log("ERROR", _error.Error()+"\n", server.Log)
		return
	}
	log("INFO", "ServeTLS: "+path+"\n", server.Log)
	server.handler(&path, &param, connection, false, customHandlerFunc)
}

func (server *ServerTLS) handler(path *string, parameter *string, connection *tls.Conn, customHandler bool,
	customHandlerFunc func(path *string, parameter *string, connection *tls.Conn)) {
	if customHandler {
		customHandlerFunc(path, parameter, connection)
	} else {
		server.defaultHandler(path, connection)
	}
}

func (server *ServerTLS) defaultHandler(path *string, connection *tls.Conn) {
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

func (server *ServerTLS) sendFile(path string, connection *tls.Conn) {
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

func (server *ServerTLS) filename(path string) string {
	cPath := filepath.Clean("/" + path)
	return server.Root + cPath
}

func (server *ServerTLS) parseURL(reader io.Reader) (string, string, error) {
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
