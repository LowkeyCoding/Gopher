package Listing

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"../Entry"
)

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

type Listing struct {
	entries []Entry.Entry
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

func (listing *Listing) AppendDir(name, path, root, host, port string) error {
	if len(listing.entries) == 0 {
		listing.entries = append(listing.entries, Entry.Entry{}) // sentinel value
		return nil
	}

	listing.entries = append(listing.entries, Entry.Entry{'1', name, path[len(root)-1:], host, port})

	return filepath.SkipDir
}

func (listing *Listing) AppendFile(name, path, root, host, port string) {
	_type := byte('9') // Binary

	for extension, gopherType := range Extensions {
		if strings.HasSuffix(path, "."+extension) {
			_type = gopherType
			break
		}
	}
	listing.entries = append(listing.entries, Entry.Entry{_type, name, path[len(root)-1:], host, port})
}
