package onramp

import (
	"context"
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/transport"
	"github.com/libp2p/go-yamux/v4"

	//"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
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
	net.Listener
	*yamux.Session
	network.ConnectionState
	Garlic
	string
	open bool
}

// CloseRead implements network.MuxedStream.
func (*GarlicConn) CloseRead() error {
	return fmt.Errorf("CloseRead is unimplemented")
}

// CloseWrite implements network.MuxedStream.
func (*GarlicConn) CloseWrite() error {
	return fmt.Errorf("CloseWrite is unimplemented")
}

// Reset implements network.MuxedStream.
func (*GarlicConn) Reset() error {
	return fmt.Errorf("Reset is unimplemented")
}

// Close implements transport.CapableConn.
func (g *GarlicConn) Close() error {
	g.Conn.Close()
	g.Listener.Close()
	g.Garlic.Close()
	g.open = false
	return nil
}

// ConnState implements transport.CapableConn.
func (g *GarlicConn) ConnState() network.ConnectionState {
	return g.ConnectionState
}

// IsClosed implements transport.CapableConn.
func (g *GarlicConn) IsClosed() bool {
	return !g.open
}

// LocalMultiaddr implements transport.CapableConn.
func (*GarlicConn) LocalMultiaddr() ma.Multiaddr {
	panic("LocalMultiaddr is unimplemented")
}

// LocalPeer implements transport.CapableConn.
func (*GarlicConn) LocalPeer() peer.ID {
	panic("LocalPeer is unimplemented")
}

// OpenStream implements transport.CapableConn.
func (*GarlicConn) OpenStream(context.Context) (network.MuxedStream, error) {
	return nil, fmt.Errorf("OpenStream is unimplemented")
}

// RemoteMultiaddr implements transport.CapableConn.
func (*GarlicConn) RemoteMultiaddr() ma.Multiaddr {
	panic("RemoteMultiaddr is unimplemented")
}

// RemotePeer implements transport.CapableConn.
func (*GarlicConn) RemotePeer() peer.ID {
	panic("RemotePeer is unimplemented")
}

// RemotePublicKey implements transport.CapableConn.
func (*GarlicConn) RemotePublicKey() crypto.PubKey {
	panic("RemotePublicKey is unimplemented")
}

// Scope implements transport.CapableConn.
func (*GarlicConn) Scope() network.ConnScope {
	panic("Scope is unimplemented")
}

// Transport implements transport.CapableConn.
func (g *GarlicConn) Transport() transport.Transport {
	return &GarlicP2P{
		garlic: g,
	}
}

func (g *GarlicConn) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (transport.CapableConn, error) {
	g.ConnectionState = network.ConnectionState{
		StreamMultiplexer:         "",
		Security:                  "",
		Transport:                 "garlic",
		UsedEarlyMuxerNegotiation: false,
	}
	addr, err := raddr.ValueForProtocol(ma.P_GARLIC32)
	if err != nil {
		return nil, err
	}
	g.Conn, err = g.Garlic.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (g *GarlicConn) AcceptStream() (network.MuxedStream, error) {
	stream, err := g.Listener.Accept()
	if err != nil {
		return nil, err
	}
	g.Session, err = yamux.Server(stream, nil, nil)
	if err != nil {
		return nil, err
	}
	return g, nil
}

// GarlicP2P Implements the Transport interface from libp2p by wrapping a GarlicConn struct
type GarlicP2P struct {
	garlic        *GarlicConn
	ClientSession *yamux.Session
	ServerSession *yamux.Session
}

// CanDial implements transport.Transport.
func (*GarlicP2P) CanDial(addr ma.Multiaddr) bool {
	panic("unimplemented: CanDial")
}

// Listen implements transport.Transport.
func (*GarlicP2P) Listen(laddr ma.Multiaddr) (transport.Listener, error) {
	return nil, fmt.Errorf("unimplemented: Listen")
}

// Protocols implements transport.Transport.
func (*GarlicP2P) Protocols() []int {
	panic("unimplemented: Protocols")
}

// Proxy implements transport.Transport.
func (*GarlicP2P) Proxy() bool {
	panic("unimplemented: Proxy")
}

var tport transport.Transport = &GarlicP2P{}

func (g *GarlicP2P) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (transport.CapableConn, error) {
	return g.garlic.Dial(ctx, raddr, p)
}
