package entry

import "fmt"

//Entry is a struct to contain values for canonical types
type Entry struct {
	Type     byte
	Display  string
	Selector string
	Hostname string
	Port     string
}

//Formats a string to fit any canonical type
func (entry Entry) String() string {
	return fmt.Sprintf("%c%s\t%s\t%s\t%s\r\n",
		entry.Type, entry.Display, entry.Selector, entry.Hostname, entry.Port)
}
