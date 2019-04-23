package Entry

import "fmt"

type Entry struct {
	Type     byte
	Display  string
	Selector string
	Hostname string
	Port     string
}

func (entry Entry) String() string {
	return fmt.Sprintf("%c%s\t%s\t%s\t%s\r\n",
		entry.Type, entry.Display, entry.Selector, entry.Hostname, entry.Port)
}
