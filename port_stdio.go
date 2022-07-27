package jsonrpc

import (
	"bufio"
	"os"
	"sync"
)

type stdioListener struct {
}

func NewStdioListener() Listener {
	return &stdioListener{}
}

func (l stdioListener) Close() {

}

func (l stdioListener) Accept() (Port, error) {
	return NewStdioPort(), nil
}

type stdioPort struct {
	file   *os.File
	rd     *bufio.Reader
	wr     *bufio.Writer
	wrLock sync.Mutex
}

func NewStdioPort() Port {
	return &stdioPort{
		file: os.Stdin,
		rd:   bufio.NewReader(os.Stdin),
		wr:   bufio.NewWriter(os.Stdout),
	}
}

func (p *stdioPort) Close() {
	p.file.Close()
}

func (p *stdioPort) Receive() (*HeaderSet, []byte, error) {
	if hdrs, bytes, err := ReceiveFromBuffered(p.rd); err != nil {
		if err == os.ErrClosed {
			return nil, nil, nil
		}
		return hdrs, nil, err
	} else {
		return hdrs, bytes, err
	}
}

func (p *stdioPort) Send(hdrs *HeaderSet, bytes []byte) error {
	p.wrLock.Lock()
	defer p.wrLock.Unlock()

	return SendToBuffered(p.wr, hdrs, bytes)
}
