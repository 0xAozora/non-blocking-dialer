//go:build windows && cgo
// +build windows,cgo

package nb_dialer

//#cgo windows LDFLAGS: -lws2_32 -lwsock32
//#include"nonblock.h"
import "C"
import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

type WinConn struct {
	net.Conn // Wrap this just to satisfy the interface
	fd       windows.Handle
	addr     *windows.SockaddrInet4
}

func (c *WinConn) Read(b []byte) (int, error) {
	n, _, err := windows.Recvfrom(c.fd, b, 0)
	if err != nil && err == syscall.EAFNOSUPPORT {
		err = nil
	}
	return n, err
}

func (c *WinConn) Write(b []byte) (int, error) {
	return 0, windows.Sendto(c.fd, b, 0, c.addr)
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

	if network != "tcp" {
		return nil, errors.New("unsupported network type")
	}

	fd, err := windows.Socket(windows.AF_INET, windows.SOCK_STREAM, windows.IPPROTO_TCP)
	if err != nil {
		return nil, err
	}

	// Set NonBlocking
	ok := C.set_nonblock(C.int(fd), C.bool(true))
	if !bool(ok) {
		return nil, errors.New("set_nonblock error")
	}

	addr, err := parseSockaddrInet4(address)
	if err != nil {
		return nil, err
	}

	err = windows.Connect(fd, addr)
	if err != windows.WSAEWOULDBLOCK {
		return nil, err
	}

	return &WinConn{fd: fd, addr: addr}, nil
}

func parseSockaddrInet4(address string) (*windows.SockaddrInet4, error) {
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid address format: %s", address)
	}

	ip := net.ParseIP(parts[0]).To4()
	if ip == nil {
		return nil, fmt.Errorf("invalid IPv4 address: %s", parts[0])
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %s", parts[1])
	}

	sa := &windows.SockaddrInet4{
		Port: port,
	}
	copy(sa.Addr[:], ip)

	return sa, nil
}
