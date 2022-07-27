package jsonrpc

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

type Header struct {
	Name  string
	Value string
}

func NewHeader(name string, value string) *Header {
	return &Header{
		Name:  name,
		Value: value,
	}
}

type HeaderSet struct {
	items []*Header
}

func NewHeaderSet() *HeaderSet {
	return &HeaderSet{
		items: make([]*Header, 0),
	}
}

func (s *HeaderSet) Get(name string) *Header {
	for i := 0; i < len(s.items); i++ {
		if strings.EqualFold(s.items[i].Name, name) {
			return s.items[i]
		}
	}

	return nil
}

func (s *HeaderSet) Set(name string, value string) {
	var hdr *Header

	if hdr = s.Get(name); hdr != nil {
		hdr.Value = value
	} else {
		s.Add(NewHeader(name, value))
	}
}

func (s *HeaderSet) Add(item *Header) {
	s.items = append(s.items, item)
}

func (s *HeaderSet) Clear() {
	s.items = make([]*Header, 0)
}

func (s *HeaderSet) ContentLength() (int, error) {
	if hdr := s.Get(contentLengthHeaderName); hdr != nil {
		if val, err := strconv.Atoi(hdr.Value); err != nil {
			return 0, err
		} else {
			if val < 0 {
				return 0, fmt.Errorf("content length must not be smaller than 0")
			}

			return val, nil
		}
	}

	return 0, fmt.Errorf("'%s' header is missing", contentLengthHeaderName)
}

func (s *HeaderSet) SetContentLength(value int) {
	s.Set(contentLengthHeaderName, strconv.Itoa(value))
}

func (s *HeaderSet) WriteTo(wr *bufio.Writer) error {
	for i := 0; i < len(s.items); i++ {
		hdr := s.items[i]

		if err := writeLine(wr, hdr.Name+": "+hdr.Value); err != nil {
			return err
		}
	}

	if err := writeLine(wr, ""); err != nil {
		return err
	}

	return nil
}

func ReadHeaderSet(rd *bufio.Reader) (*HeaderSet, error) {
	var line string
	var err error

	res := NewHeaderSet()

	if line, err = readLine(rd); err != nil {
		return nil, err
	}

	for len(line) > 0 {
		var idx int

		if idx = strings.IndexByte(line, ':'); idx < 0 {
			return nil, fmt.Errorf("invalid header; missing ':' name-value separator")
		}

		name := strings.TrimSpace(line[0:idx])
		value := strings.TrimSpace(line[idx+1:])

		res.Add(NewHeader(name, value))

		if line, err = readLine(rd); err != nil {
			return nil, err
		}
	}

	return res, nil
}

const crlf string = "\r\n"
const contentLengthHeaderName string = "Content-Length"

func readLine(rd *bufio.Reader) (string, error) {
	if line, err := rd.ReadString('\n'); err != nil {
		return "", err
	} else {
		return strings.TrimSuffix(line, crlf), nil
	}
}

func writeLine(wr *bufio.Writer, line string) error {
	if _, err := wr.WriteString(line + crlf); err != nil {
		return err
	}

	return nil
}
