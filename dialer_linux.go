//go:build linux
// +build linux

package nb_dialer

import (
	"net"
	"os"
	"syscall"
)

type Conn struct {
	net.Conn // Wrap this just to satisfy the interface
	fd       uint64
}

func (c *Conn) GetFD() uint64 {
	return c.fd
}

func (c *Conn) Validate() (net.Conn, error) {

	_, err := syscall.GetsockoptInt(int(c.fd), syscall.SOL_SOCKET, syscall.SO_ERROR)
	if err != nil {
		return nil, err
	}

	file := os.NewFile(uintptr(c.fd), "")
	conn, err := net.FileConn(file)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *Conn) SetNonBlock(nonblock bool) error {
	return syscall.SetNonblock(int(c.fd), nonblock)
}

type Dialer struct{}

func (d *Dialer) Dial(network string, address string) (NonBlockConn, error) {

	net.Dial(network, address)

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return nil, err
	}

	// Set NonBlocking
	err = syscall.SetNonblock(fd, true)
	if err != nil {
		return nil, err
	}

	addr := syscall.SockaddrInet4{
		Port: 27017,
		Addr: [4]byte{205, 196, 6, 214},
	}

	err = syscall.Connect(fd, &addr)
	if err != syscall.EINPROGRESS {
		return nil, err
	}

	return &Conn{fd: uint64(fd)}, nil
}
