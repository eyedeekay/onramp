package onramp

import (
	"net"
	"time"
)

type NullConn struct {
	net.Conn
}

func (nc *NullConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (nc *NullConn) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (nc *NullConn) Close() error { return nil }

func (nc *NullConn) LocalAddr() net.Addr {
	if nc.Conn != nil {
		return nc.Conn.LocalAddr()
	}
	return &net.IPAddr{
		IP: net.ParseIP("127.0.0.1"),
	}
}

func (nc *NullConn) RemoteAddr() net.Addr {
	if nc.Conn != nil {
		return nc.Conn.RemoteAddr()
	}
	return &net.IPAddr{
		IP: net.ParseIP("127.0.0.1"),
	}
}

func (nc *NullConn) SetDeadline(t time.Time) error { return nil }

func (nc *NullConn) SetReadDeadline(t time.Time) error { return nil }

func (nc *NullConn) SetWriteDeadline(t time.Time) error { return nil }
