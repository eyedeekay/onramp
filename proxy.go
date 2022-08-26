package onramp

import (
	"io"
	"log"
	"net"
	"strings"
)

type OnrampProxy struct {
	Onion
	Garlic
}

// Proxy passes requests from a net.Listener to a remote server
// without touching them in any way. It can be used as a shortcut,
// set up a Garlic or Onion Listener and pass it, along with the
// address of a locally running service and the hidden service
// listener will expose the local service.
// Pass it a regular net.Listener(or a TLS listener if you like),
// and an I2P or Onion address, and it will act as a tunnel to a
// listening hidden service somewhere.
func (p *OnrampProxy) Proxy(list net.Listener, raddr string) error {
	for {
		conn, err := list.Accept()
		if err != nil {
			return err
		}
		go p.proxy(conn, raddr)
	}
}

func (p *OnrampProxy) proxy(conn net.Conn, raddr string) {
	var remote net.Conn
	var err error
	checkaddr := strings.Split(raddr, ":")[0]
	if strings.HasSuffix(checkaddr, ".i2p") {
		remote, err = p.Garlic.Dial("tcp", raddr)
	} else if strings.HasSuffix(checkaddr, ".onion") {
		remote, err = p.Onion.Dial("tcp", raddr)
	} else {
		remote, err = net.Dial("tcp", raddr)
	}
	if err != nil {
		log.Fatalf("cannot dial to remote: %v", err)
	}
	defer remote.Close()
	go io.Copy(remote, conn)
	io.Copy(conn, remote)
}

var proxy *OnrampProxy = &OnrampProxy{}

func Proxy(list net.Listener, raddr string) error {
	return proxy.Proxy(list, raddr)
}
