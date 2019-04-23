package listing

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	entry "../Entry"
)

//Extensions maps extensions to their apropriate canonical type.
var Extensions = map[string]byte{
	"aiff":     's',
	"au":       's',
	"c":        '0',
	"cfg":      '0',
	"cpp":      '0',
	"cs":       '0',
	"css":      '0',
	"csv":      '0',
	"gif":      'g',
	"go":       '0',
	"gpg":      '0',
	"h":        '0',
	"html":     'h',
	"ini":      '0',
	"java":     '0',
	"jpeg":     'I',
	"jpg":      'I',
	"js":       '0',
	"json":     '0',
	"log":      '0',
	"lua":      '0',
	"markdown": '0',
	"md":       '0',
	"mp3":      's',
	"php":      '0',
	"pl":       '0',
	"png":      'I',
	"py":       '0',
	"rb":       '0',
	"rss":      '0',
	"sh":       '0',
	"txt":      '0',
	"wav":      's',
	"xml":      '0',
}

//Listing a struct to contain a list of entries.
type Listing struct {
	entries []entry.Entry
}

func (listing Listing) String() string {
	var buffer bytes.Buffer

	for _, entry := range listing.entries {
		if entry.Type == '1' {
			fmt.Fprint(&buffer, entry)
		}
	}

	for _, entry := range listing.entries {
		if entry.Type == 0 || entry.Type == '1' {
			continue // skip sentinel value and directories
		}

		fmt.Fprint(&buffer, entry)
	}

	return buffer.String()
}

//AppendDir appends a directory to the list of entries. Returns a error used as a return value from WalkFuncs to indicate that the directory named in the call is to be skipped.
func (listing *Listing) AppendDir(name, path, root, host, port string) error {
	if len(listing.entries) == 0 {
		listing.entries = append(listing.entries, entry.Entry{}) // sentinel value
		return nil
	}

	listing.entries = append(listing.entries, entry.Entry{'1', name, path[len(root)-1:], host, port})

	return filepath.SkipDir
}

//AppendFile appends a file to the list of entries.
func (listing *Listing) AppendFile(name, path, root, host, port string) {
	_type := byte('9') // Binary

	for extension, gopherType := range Extensions {
		if strings.HasSuffix(path, "."+extension) {
			_type = gopherType
			break
		}
	}
	listing.entries = append(listing.entries, entry.Entry{_type, name, path[len(root)-1:], host, port})
}
