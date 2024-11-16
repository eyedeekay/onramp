package onramp

import (
	"io"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
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
	log.WithFields(logrus.Fields{
		"remote_address": raddr,
		"local_address":  list.Addr().String(),
	}).Debug("Starting proxy service")

	for {
		log.Debug("Waiting for incoming connection")
		conn, err := list.Accept()
		if err != nil {
			log.WithError(err).Error("Failed to accept connection")
			return err
		}

		log.WithFields(logrus.Fields{
			"local_addr":  conn.LocalAddr().String(),
			"remote_addr": conn.RemoteAddr().String(),
		}).Debug("Accepted new connection, starting proxy routine")

		go p.proxy(conn, raddr)
	}
}

func (p *OnrampProxy) proxy(conn net.Conn, raddr string) {
	log.WithFields(logrus.Fields{
		"remote_address": raddr,
		"local_addr":     conn.LocalAddr().String(),
		"remote_addr":    conn.RemoteAddr().String(),
	}).Debug("Setting up proxy connection")

	var remote net.Conn
	var err error
	checkaddr := strings.Split(raddr, ":")[0]
	if strings.HasSuffix(checkaddr, ".i2p") {
		log.Debug("Detected I2P address, using Garlic connection")
		remote, err = p.Garlic.Dial("tcp", raddr)
	} else if strings.HasSuffix(checkaddr, ".onion") {
		log.Debug("Detected Onion address, using Tor connection")
		remote, err = p.Onion.Dial("tcp", raddr)
	} else {
		log.Debug("Using standard TCP connection")
		remote, err = net.Dial("tcp", raddr)
	}
	if err != nil {
		log.WithError(err).Error("Failed to establish remote connection")
		log.Fatal("Cannot dial to remote")
	}
	defer remote.Close()

	log.WithFields(logrus.Fields{
		"local_addr":  remote.LocalAddr().String(),
		"remote_addr": remote.RemoteAddr().String(),
	}).Debug("Remote connection established, starting bidirectional copy")

	go io.Copy(remote, conn)
	io.Copy(conn, remote)
}

var proxy *OnrampProxy = &OnrampProxy{}

func Proxy(list net.Listener, raddr string) error {
	return proxy.Proxy(list, raddr)
}
