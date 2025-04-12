package nb_dialer

import "net"

type NonBlockConn interface {
	GetFD() uint64
	Validate() (net.Conn, error)
	SetNonBlock(bool) error
}
