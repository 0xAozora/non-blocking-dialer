//go:build linux || darwin || freebsd
// +build linux darwin freebsd

package nb_dialer

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
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

	if network != "tcp" {
		return nil, errors.New("unsupported network type")
	}

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return nil, err
	}

	// Set NonBlocking
	err = syscall.SetNonblock(fd, true)
	if err != nil {
		return nil, err
	}

	addr, err := parseSockaddrInet4(address)
	if err != nil {
		return nil, err
	}

	err = syscall.Connect(fd, addr)
	if err != syscall.EINPROGRESS {
		return nil, err
	}

	return &Conn{fd: uint64(fd)}, nil
}

func parseSockaddrInet4(address string) (*syscall.SockaddrInet4, error) {
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

	sa := &syscall.SockaddrInet4{
		Port: port,
	}
	copy(sa.Addr[:], ip)

	return sa, nil
}
