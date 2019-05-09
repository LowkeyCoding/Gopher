package main

import (
	"crypto/tls"
	"io"
	"os"
	"strings"
	"time"
)

func main() {
	client := Client{log: "./log/client.txt"}
	site := Site{}
	client.configure("./certs/client")
	client.connect("192.168.1.114:7070")
	client.writeToServer("/", &site)
	site.renderSite()
}

type Line struct {
	ItemType    string
	Displaytext string
	Selector    string
	Hostname    string
	Port        string
}

type Site struct {
	Lines []Line
}

type Client struct {
	log         string
	config      tls.Config
	connection  *tls.Conn
	lastPage    string
	currentLine int
}

func (client *Client) configure(path string) {
	cert, err := tls.LoadX509KeyPair(path+".pem", path+".key")
	if err != nil {
		log("server: loadkeys: ", err.Error(), client.log)
	}
	client.config = tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
}

func (client *Client) connect(adress string) {
	connection, err := tls.Dial("tcp", "192.168.1.114:7070", &client.config)
	if err != nil {
		log("client: dial: ", err.Error(), client.log)
	}
	client.connection = connection
}

func (client *Client) writeToServer(message string, site *Site) {
	defer client.connection.Close()
	n, err := io.WriteString(client.connection, message)
	if err != nil {
		log("client: write: ", err.Error(), client.log)
	}

	reply, n, err := client.connection.ReadDynamic()
	if err != nil {
		log("client: write: ", err.Error(), client.log)
	}
	site.parseSite(string(reply[:n]))
}

func (site *Site) parseSite(reply string) {
	siteSlice := strings.Split(reply, "\n")
	for _, line := range siteSlice {
		line := strings.Split(line, "\t")
		if len(line[0]) <= 1 {
			continue
		} else {
			_ItemType := string(string(line[0])[0])
			if _ItemType == "i" {
				_Displaytext := string(string(line[0])[1:])
				_Selector := ""
				_Hostname := ""
				_Port := ""
				_Line := Line{_ItemType, _Displaytext, _Selector, _Hostname, _Port}
				site.Lines = append(site.Lines, _Line)
			} else {
				_Displaytext := string(string(line[0])[1:])
				_Selector := string(line[1])
				_Hostname := string(line[2])
				_Port := string(line[3])
				_Line := Line{_ItemType, _Displaytext, _Selector, _Hostname, _Port}
				site.Lines = append(site.Lines, _Line)
			}
		}
	}
}

func (site *Site) renderSite() {
	for _, line := range site.Lines {
		_prefix := line.resolvePrefix()
		if _prefix == "" {
			println(line.Displaytext)
		} else {
			println(_prefix + " " + line.Displaytext)
		}
	}
}

func (line *Line) resolvePrefix() string {
	switch line.ItemType {
	case "0":
		return "<TXT>"
	case "1":
		return "<DIR>"
	case "2":
		return "<CCSO>"
	case "3":
		return "<ERROR>"
	case "4":
		return "<BINHEX>"
	case "5":
		return "<DOS>"
	case "6":
		return "<UU>"
	case "7":
		return "<SEARCH>"
	case "8":
		return "<TEL>"
	case "9":
		return "<BIN>"
	case "+":
		return "<MIRROR>"
	case "g":
		return "<GIF>"
	case "I":
		return "<IMG>"
	case "T":
		return "<TEL3270>"
	case "h":
		return "<HTML>"
	case "i":
		return ""
	case "s":
		return "<SOUND>"
	}
	return ""
}

func log(infoType string, info string, log string) {
	currentTime := "[" + time.Now().Format("2006-01-02 15:04:05") + "] "
	println(currentTime + "[" + infoType + "] " + info)
	file, _error := os.OpenFile(log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if _error != nil {
		println(currentTime + "[ERROR] " + _error.Error() + "\n")
		return
	}
	if _, _error := file.Write([]byte(currentTime + "[" + infoType + "] " + info)); _error != nil {
		println(currentTime + "[ERROR] " + _error.Error() + "\n")
		return
	}
	if _error := file.Close(); _error != nil {
		println(currentTime + "[ERROR] " + _error.Error() + "\n")
		return
	}
	return
}
