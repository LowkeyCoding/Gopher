package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell"
)

func main() {
	site := Site{}
	client := Client{log: "./log/client.txt", currentPage: &site, currentLine: 0, cursorPos: 0, scrollOffset: 0}
	client.configure("./certs/client")
	client.connect("192.168.1.114:7070")
	client.writeToServer("/", &site)
	client.visitedPages = append(client.visitedPages, Line{"1", "index", "/", "192.168.1.114", "7070"})
	screen := client.initRendere()
	screen.Clear()
	site.renderSite(screen, &client)
	client.rendere(screen)
	screen.Show()
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
	log          string
	config       tls.Config
	connection   *tls.Conn
	currentPage  *Site
	visitedPages []Line
	currentLine  int
	cursorPos    int
	lineOffset   int
	scrollOffset int
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

func (client *Client) initRendere() tcell.Screen {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	screen, err := tcell.NewScreen()
	if err != nil {
		log("client: write: ", err.Error(), client.log)
	}
	if err = screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	screen.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack))
	screen.Clear()

	return screen
}

func (client *Client) rendere(screen tcell.Screen) {
	quit := make(chan struct{})
	client.currentPage.renderSite(screen, client)
	go func() {
		for {
			event := screen.PollEvent()
			switch event := event.(type) {
			case *tcell.EventKey:
				switch event.Key() {
				case tcell.KeyUp:
					client.currentLine--
					client.cursorPos--
					_, height := screen.Size()
					if client.cursorPos < 0 && len(client.currentPage.Lines) < 15 {
						client.cursorPos = len(client.currentPage.Lines)
					} else if client.cursorPos < 0 {
						client.cursorPos = 0
						client.scrollOffset--
						if client.scrollOffset < 0 {
							client.scrollOffset = len(client.currentPage.Lines) - height + 1
							client.cursorPos = height - 1
							client.currentLine = len(client.currentPage.Lines) - 1
						}
						screen.Clear()
						client.currentPage.renderSite(screen, client)
					}
				case tcell.KeyDown:
					client.currentLine++
					client.cursorPos++
					_, height := screen.Size()
					if client.cursorPos > len(client.currentPage.Lines)-1 {
						client.cursorPos = 0
					} else if client.cursorPos >= height {
						client.scrollOffset++
						client.cursorPos = height - 1
						if client.scrollOffset > len(client.currentPage.Lines)-height+1 {
							client.scrollOffset = 0
							client.cursorPos = 0
							client.currentLine = 0
							screen.Clear()
							client.currentPage.renderSite(screen, client)
						}
						screen.Clear()
						client.currentPage.renderSite(screen, client)
					}
				case tcell.KeyEnter:
					if client.currentLine > 1 {
						line := client.currentPage.Lines[client.currentLine]
						log("DEBUG", "ItemType: "+line.ItemType+" DisplayText: "+line.Displaytext, "./log/client.txt")
						if line.ItemType == "1" {
							client.connect(line.Hostname + ":" + line.Port)
							client.writeToServer(line.Selector, client.currentPage)
							screen.Clear()
							client.currentPage.renderSite(screen, client)
							client.visitedPages = append(client.visitedPages, line)
							client.currentLine = 0
							client.cursorPos = 0
						}
					}
				case tcell.KeyCtrlL:
					screen.Sync()
				case tcell.KeyEscape:
					close(quit)
					return
				case tcell.KeyBackspace:
					if len(client.visitedPages) > 1 {
						line := client.visitedPages[len(client.visitedPages)-2]
						client.connect(line.Hostname + ":" + line.Port)
						client.writeToServer(line.Selector, client.currentPage)
						screen.Clear()
						client.visitedPages = client.visitedPages[:len(client.visitedPages)-1]
						client.currentLine = 0
						client.cursorPos = 0
						client.scrollOffset = 0
						client.currentPage.renderSite(screen, client)
					}
				}
			case *tcell.EventResize:
				screen.Sync()
			}
			screen.ShowCursor(0, client.cursorPos)
		}
	}()

loop:
	for {
		select {
		case <-quit:
			break loop
		case <-time.After(time.Millisecond):
		}

	}

	screen.Fini()

}

func (site *Site) parseSite(reply string) {
	siteSlice := strings.Split(reply, "\n")
	site.Lines = nil
	for _, line := range siteSlice {
		line := strings.Split(line, "\t")
		if len(line[0]) <= 1 {
			continue
		} else {
			_ItemType := string(string(line[0])[0])
			if _ItemType == "i" {
				_Displaytext := addToString(string(string(line[0])[1:]), "\t", (8%len(string(string(line[0])[1:])))+1)
				_Selector := ""
				_Hostname := ""
				_Port := ""
				_Line := Line{_ItemType, _Displaytext, _Selector, _Hostname, _Port}
				site.Lines = append(site.Lines, _Line)
			} else {
				_Displaytext := addToString(string(string(line[0])[1:]), "\t", (8%len(string(string(line[0])[1:])))+1)
				_Selector := string(line[1])
				_Hostname := string(line[2])
				_Port := string(line[3])
				_Line := Line{_ItemType, _Displaytext, _Selector, _Hostname, _Port}
				site.Lines = append(site.Lines, _Line)
			}
		}
	}
}

func (site *Site) renderSite(screen tcell.Screen, client *Client) {
	style := tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack)
	for y, line := range site.Lines[client.scrollOffset:] {
		_prefix := line.resolvePrefix()
		if _prefix == "" {
			for x, char := range line.Displaytext {
				screen.SetCell(x, y, style, char)
			}
		} else {
			for x, char := range _prefix + " " + line.Displaytext {
				screen.SetCell(x, y, style, char)
			}
		}
	}
	screen.Show()
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

func addToString(initialString, addString string, times int) string {
	for i := 0; i < times; i++ {
		initialString += addString
	}
	return initialString
}

func log(infoType, info, log string) {
	currentTime := "[" + time.Now().Format("2006-01-02 15:04:05") + "] "
	println(currentTime + "[" + infoType + "] " + info + "\n")
	file, _error := os.OpenFile(log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if _error != nil {
		println(currentTime + "[ERROR] " + _error.Error() + "\n")
		return
	}
	if _, _error := file.Write([]byte(currentTime + "[" + infoType + "] " + info + "\n")); _error != nil {
		println(currentTime + "[ERROR] " + _error.Error() + "\n")
		return
	}
	if _error := file.Close(); _error != nil {
		println(currentTime + "[ERROR] " + _error.Error() + "\n")
		return
	}
	return
}
