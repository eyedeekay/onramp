package onramp

import (
	"context"
	"net"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/transport"
	ma "github.com/multiformats/go-multiaddr"
)

/*
type Transport interface {
	// Dial dials a remote peer. It should try to reuse local listener
	// addresses if possible but it may choose not to.
	Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (transport.CapableConn, error)

	// CanDial returns true if this transport knows how to dial the given
	// multiaddr.
	//
	// Returning true does not guarantee that dialing this multiaddr will
	// succeed. This function should *only* be used to preemptively filter
	// out addresses that we can't dial.
	CanDial(addr ma.Multiaddr) bool

	// Listen listens on the passed multiaddr.
	Listen(laddr ma.Multiaddr) (transport.Listener, error)

	// Protocol returns the set of protocols handled by this transport.
	//
	// See the Network interface for an explanation of how this is used.
	Protocols() []int

	// Proxy returns true if this is a proxy transport.
	//
	// See the Network interface for an explanation of how this is used.
	// TODO: Make this a part of the go-multiaddr protocol instead?
	Proxy() bool
}
*/

/*
type CapableConn interface {
	network.MuxedConn
	network.ConnSecurity
	network.ConnMultiaddrs
	network.ConnScoper

	// Transport returns the transport to which this connection belongs.
	Transport() Transport
}
*/

type GarlicConn struct {
	net.Conn
	Garlic
}

func (g *GarlicConn) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (transport.CapableConn, error) {
	addr, err := raddr.ValueForProtocol(ma.P_GARLIC32)
	g.Conn, err = g.Garlic.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (g *GarlicConn) AcceptStream() (network.MuxedStream, error)
	// TODO:
}

// GarlicP2P Implements the Transport interface from libp2p by wrapping a GarlicConn struct
type GarlicP2P struct {
	GarlicConn
}

func (g *GarlicP2P) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (transport.CapableConn, error) {
	return g.Dial(ctx, raddr, p)
}
