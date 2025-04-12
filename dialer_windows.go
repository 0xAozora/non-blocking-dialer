//go:build windows && cgo
// +build windows,cgo

package nb_dialer

//#cgo windows LDFLAGS: -lws2_32 -lwsock32
//#include"nonblock.h"
import "C"
import (
	"errors"
	"net"

	"golang.org/x/sys/windows"
)

type WinConn struct {
	net.Conn // Wrap this just to satisfy the interface
	fd       windows.Handle
}

func (c *WinConn) Read(b []byte) (int, error) {
	return windows.Read(c.fd, b)
}

func (c *WinConn) Write(b []byte) (int, error) {
	return windows.Write(c.fd, b)
}

func (c *WinConn) Close() error {
	return windows.Closesocket(c.fd)
}

func (c *WinConn) GetFD() uint64 {
	return uint64(c.fd)
}

func (c *WinConn) Validate() (net.Conn, error) {

	// TODO: Check if the socket is still valid
	// _, err := syscall.GetsockoptInt(syscall.Handle(c.fd), syscall.SOL_SOCKET, syscall.SO_ERROR)
	// if err != nil {
	// 	return nil, err
	// }

	// Unfortunatelly you cannot create a net.Conn from a windows.Handle
	// https://github.com/golang/go/issues/9503

	return c, nil
}

func (c *WinConn) SetNonBlock(nonblock bool) error {
	ok := C.set_nonblock(C.int(c.fd), C.bool(nonblock))
	if !bool(ok) {
		return errors.New("set_nonblock error")
	}
	return nil
}

type Dialer struct{}

func (d *Dialer) Dial(network string, address string) (NonBlockConn, error) {

	fd, err := windows.Socket(windows.AF_INET, windows.SOCK_STREAM, windows.IPPROTO_TCP)
	if err != nil {
		return nil, err
	}

	// Set NonBlocking
	ok := C.set_nonblock(C.int(fd), C.bool(true))
	if !bool(ok) {
		return nil, errors.New("set_nonblock error")
	}

	addr := windows.SockaddrInet4{
		Port: 27017,
		Addr: [4]byte{205, 196, 6, 214},
	}

	err = windows.Connect(fd, &addr)
	if err != windows.WSAEWOULDBLOCK {
		return nil, err
	}

	return &WinConn{fd: fd}, nil
}
