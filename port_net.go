package jsonrpc

import (
	"bufio"
	"net"
	"sync"
)

type netListener struct {
	net net.Listener
}

func NewNetListener(network string, address string) (Listener, error) {
	var lis net.Listener
	var err error

	if lis, err = net.Listen(network, address); err != nil {
		return nil, err
	}

	return netListener{
		net: lis,
	}, nil
}

func NewTcpListener(address string) (Listener, error) {
	return NewNetListener("tcp", address)
}

func (l netListener) Close() {
	l.net.Close()
}

func (l netListener) Accept() (Port, error) {
	if conn, err := l.net.Accept(); err != nil {
		return nil, err
	} else {
		return NewNetPort(conn), nil
	}
}

type netPort struct {
	net    net.Conn
	rd     *bufio.Reader
	wr     *bufio.Writer
	wrLock sync.Mutex
}

func NewNetPort(conn net.Conn) Port {
	return &netPort{
		net: conn,
		rd:  bufio.NewReader(conn),
		wr:  bufio.NewWriter(conn),
	}
}

func (p *netPort) Close() {
	p.net.Close()
}

func (p *netPort) Receive() (*HeaderSet, []byte, error) {
	if hdrs, bytes, err := ReceiveFromBuffered(p.rd); err != nil {
		if err == net.ErrClosed {
			return nil, nil, nil
		}

		return hdrs, nil, err
	} else {
		return hdrs, bytes, err
	}
}

func (p *netPort) Send(hdrs *HeaderSet, bytes []byte) error {
	p.wrLock.Lock()
	defer p.wrLock.Unlock()

	return SendToBuffered(p.wr, hdrs, bytes)
}
