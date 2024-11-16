//go:build !gen
// +build !gen

package onramp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/go-i2p/i2pkeys"
	"github.com/go-i2p/sam3"
)

// Garlic is a ready-made I2P streaming manager. Once initialized it always
// has a valid I2PKeys and StreamSession.
type Garlic struct {
	*sam3.StreamListener
	*sam3.StreamSession
	*sam3.DatagramSession
	ServiceKeys *i2pkeys.I2PKeys
	*sam3.SAM
	name        string
	addr        string
	opts        []string
	AddrMode    int
	TorrentMode bool
}

const (
	DEST_BASE32           = 0
	DEST_HASH             = 1
	DEST_HASH_BYTES       = 2
	DEST_BASE32_TRUNCATED = 3
	DEST_BASE64           = 4
	DEST_BASE64_BYTES     = 5
)

func (g *Garlic) Network() string {
	if g.StreamListener != nil {
		return "tcp"
	} else {
		return "udp"
	}
}

func (g *Garlic) addrString(addr string) string {
	if g.TorrentMode {
		return addr + ".i2p"
	}
	return addr
}

func (g *Garlic) String() string {
	var r string
	switch g.AddrMode {
	case DEST_HASH:
		r = g.ServiceKeys.Address.DestHash().Hash()
	case DEST_HASH_BYTES:
		hash := g.ServiceKeys.Address.DestHash()
		r = string(hash[:])
	case DEST_BASE32_TRUNCATED:
		r = strings.TrimSuffix(g.ServiceKeys.Address.Base32(), ".b32.i2p")
	case DEST_BASE32:
		r = g.ServiceKeys.Address.Base32()
	case DEST_BASE64:
		r = g.ServiceKeys.Address.Base64()
	case DEST_BASE64_BYTES:
		r = string(g.ServiceKeys.Address.Bytes())
	default:
		r = g.ServiceKeys.Address.DestHash().Hash()
	}
	return g.addrString(r) // r //strings.TrimLeft(strings.TrimRight(r, "\n"), "\n") //strings.TrimSpace(r)
}

func (g *Garlic) getName() string {
	if g.name == "" {
		return "onramp-garlic"
	}
	return g.name
}

func (g *Garlic) getAddr() string {
	if g.addr == "" {
		return "localhost:7656"
	}
	return g.addr
}

func (g *Garlic) getOptions() []string {
	if g.opts == nil {
		return OPT_DEFAULTS
	}
	return g.opts
}

func (g *Garlic) samSession() (*sam3.SAM, error) {
	if g.SAM == nil {
		log.WithField("address", g.getAddr()).Debug("Creating new SAM session")
		var err error
		g.SAM, err = sam3.NewSAM(g.getAddr())
		if err != nil {
			log.WithError(err).Error("Failed to create SAM session")
			return nil, fmt.Errorf("onramp samSession: %v", err)
		}
		log.Debug("SAM session created successfully")
	}
	return g.SAM, nil
}

func (g *Garlic) setupStreamSession() (*sam3.StreamSession, error) {
	if g.StreamSession == nil {
		log.WithField("name", g.getName()).Debug("Setting up stream session")
		var err error
		g.ServiceKeys, err = g.Keys()
		if err != nil {
			log.WithError(err).Error("Failed to get keys for stream session")
			return nil, fmt.Errorf("onramp setupStreamSession: %v", err)
		}
		log.WithField("address", g.ServiceKeys.Address.Base32()).Debug("Creating stream session with keys")
		log.Println("Creating stream session with keys:", g.ServiceKeys.Address.Base32())
		g.StreamSession, err = g.SAM.NewStreamSession(g.getName(), *g.ServiceKeys, g.getOptions())
		if err != nil {
			log.WithError(err).Error("Failed to create stream session")
			return nil, fmt.Errorf("onramp setupStreamSession: %v", err)
		}
		log.Debug("Stream session created successfully")
		return g.StreamSession, nil
	}
	return g.StreamSession, nil
}

func (g *Garlic) setupDatagramSession() (*sam3.DatagramSession, error) {
	if g.DatagramSession == nil {
		log.WithField("name", g.getName()).Debug("Setting up datagram session")
		var err error
		g.ServiceKeys, err = g.Keys()
		if err != nil {
			log.WithError(err).Error("Failed to get keys for datagram session")
			return nil, fmt.Errorf("onramp setupDatagramSession: %v", err)
		}
		log.WithField("address", g.ServiceKeys.Address.Base32()).Debug("Creating datagram session with keys")
		log.Println("Creating datagram session with keys:", g.ServiceKeys.Address.Base32())
		g.DatagramSession, err = g.SAM.NewDatagramSession(g.getName(), *g.ServiceKeys, g.getOptions(), 0)
		if err != nil {
			log.WithError(err).Error("Failed to create datagram session")
			return nil, fmt.Errorf("onramp setupDatagramSession: %v", err)
		}
		log.Debug("Datagram session created successfully")
		return g.DatagramSession, nil
	}

	log.Debug("Using existing datagram session")
	return g.DatagramSession, nil
}

// NewListener returns a net.Listener for the Garlic structure's I2P keys.
// accepts a variable list of arguments, arguments after the first one are ignored.
func (g *Garlic) NewListener(n, addr string) (net.Listener, error) {
	log.WithFields(logrus.Fields{
		"network": n,
		"address": addr,
		"name":    g.getName(),
	}).Debug("Creating new listener")
	listener, err := g.Listen(n)
	if err != nil {
		log.WithError(err).Error("Failed to create listener")
		return nil, err
	}

	log.Debug("Successfully created listener")
	return listener, nil
	// return g.Listen(n)
}

// Listen returns a net.Listener for the Garlic structure's I2P keys.
// accepts a variable list of arguments, arguments after the first one are ignored.
func (g *Garlic) Listen(args ...string) (net.Listener, error) {
	log.WithFields(logrus.Fields{
		"args": args,
		"name": g.getName(),
	}).Debug("Setting up listener")

	listener, err := g.OldListen(args...)
	if err != nil {
		log.WithError(err).Error("Failed to create listener")
		return nil, err
	}

	log.Debug("Successfully created listener")
	return listener, nil
	// return g.OldListen(args...)
}

// OldListen returns a net.Listener for the Garlic structure's I2P keys.
// accepts a variable list of arguments, arguments after the first one are ignored.
func (g *Garlic) OldListen(args ...string) (net.Listener, error) {
	log.WithField("args", args).Debug("Starting OldListen")
	if len(args) > 0 {
		protocol := args[0]
		log.WithField("protocol", protocol).Debug("Checking protocol type")
		// if args[0] == "tcp" || args[0] == "tcp6" || args[0] == "st" || args[0] == "st6" {
		if protocol == "tcp" || protocol == "tcp6" || protocol == "st" || protocol == "st6" {
			log.Debug("Using TCP stream listener")
			return g.ListenStream()
			//} else if args[0] == "udp" || args[0] == "udp6" || args[0] == "dg" || args[0] == "dg6" {
		} else if protocol == "udp" || protocol == "udp6" || protocol == "dg" || protocol == "dg6" {
			log.Debug("Using UDP datagram listener")
			pk, err := g.ListenPacket()
			if err != nil {
				log.WithError(err).Error("Failed to create packet listener")
				return nil, err
			}
			log.Debug("Successfully created datagram session")
			return pk.(*sam3.DatagramSession), nil
		}

	}
	log.Debug("No protocol specified, defaulting to stream listener")
	return g.ListenStream()
}

// Listen returns a net.Listener for the Garlic structure's I2P keys.
func (g *Garlic) ListenStream() (net.Listener, error) {
	log.Debug("Setting up stream listener")
	var err error
	if g.SAM, err = g.samSession(); err != nil {
		log.WithError(err).Error("Failed to create SAM session for stream listener")
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	if g.StreamSession, err = g.setupStreamSession(); err != nil {
		log.WithError(err).Error("Failed to setup stream session")
		return nil, fmt.Errorf("onramp Listen: %v", err)
	}
	if g.StreamListener == nil {
		log.Debug("Creating new stream listener")
		g.StreamListener, err = g.StreamSession.Listen()
		if err != nil {
			log.WithError(err).Error("Failed to create stream listener")
			return nil, fmt.Errorf("onramp Listen: %v", err)
		}
		log.Debug("Stream listener created successfully")
	}
	return g.StreamListener, nil
}

// ListenPacket returns a net.PacketConn for the Garlic structure's I2P keys.
func (g *Garlic) ListenPacket() (net.PacketConn, error) {
	log.Debug("Setting up packet connection")
	var err error
	if g.SAM, err = g.samSession(); err != nil {
		log.WithError(err).Error("Failed to create SAM session for packet connection")
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	if g.DatagramSession, err = g.setupDatagramSession(); err != nil {
		log.WithError(err).Error("Failed to setup datagram session")
		return nil, fmt.Errorf("onramp Listen: %v", err)
	}
	log.Debug("Packet connection successfully established")
	return g.DatagramSession, nil
}

// ListenTLS returns a net.Listener for the Garlic structure's I2P keys,
// which also uses TLS either for additional encryption, authentication,
// or browser-compatibility.
func (g *Garlic) ListenTLS(args ...string) (net.Listener, error) {
	log.WithField("args", args).Debug("Starting TLS listener")
	listener, err := g.Listen(args...)
	if err != nil {
		log.WithError(err).Error("Failed to create base listener")
		return nil, err
	}
	cert, err := g.TLSKeys()
	if err != nil {
		log.WithError(err).Error("Failed to get TLS keys")
		return nil, fmt.Errorf("onramp ListenTLS: %v", err)
	}
	if len(args) > 0 {
		protocol := args[0]
		log.WithField("protocol", protocol).Debug("Creating TLS listener for protocol")

		// if args[0] == "tcp" || args[0] == "tcp6" || args[0] == "st" || args[0] == "st6" {
		if protocol == "tcp" || protocol == "tcp6" || protocol == "st" || protocol == "st6" {
			log.Debug("Creating TLS stream listener")
			return tls.NewListener(
				g.StreamListener,
				&tls.Config{
					Certificates: []tls.Certificate{cert},
				},
			), nil
			//} else if args[0] == "udp" || args[0] == "udp6" || args[0] == "dg" || args[0] == "dg6" {
		} else if protocol == "udp" || protocol == "udp6" || protocol == "dg" || protocol == "dg6" {
			log.Debug("Creating TLS datagram listener")
			return tls.NewListener(
				g.DatagramSession,
				&tls.Config{
					Certificates: []tls.Certificate{cert},
				},
			), nil
		}

	} else {
		log.Debug("No protocol specified, using stream listener")
		g.StreamListener = listener.(*sam3.StreamListener)
	}
	log.Debug("Successfully created TLS listener")
	return tls.NewListener(
		g.StreamListener,
		&tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	), nil
}

// Dial returns a net.Conn for the Garlic structure's I2P keys.
func (g *Garlic) Dial(net, addr string) (net.Conn, error) {
	log.WithFields(logrus.Fields{
		"network": net,
		"address": addr,
	}).Debug("Attempting to dial")
	if !strings.Contains(addr, ".i2p") {
		log.Debug("Non-I2P address detected, returning null connection")
		return &NullConn{}, nil
	}
	var err error
	if g.SAM, err = g.samSession(); err != nil {
		log.WithError(err).Error("Failed to create SAM session")
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	if g.StreamSession, err = g.setupStreamSession(); err != nil {
		log.WithError(err).Error("Failed to setup stream session")
		return nil, fmt.Errorf("onramp Dial: %v", err)
	}
	log.Debug("Attempting to establish connection")
	conn, err := g.StreamSession.Dial(net, addr)
	if err != nil {
		log.WithError(err).Error("Failed to establish connection")
		return nil, err
	}
	log.Debug("Successfully established connection")
	return conn, nil
	// return g.StreamSession.Dial(net, addr)
}

// DialContext returns a net.Conn for the Garlic structure's I2P keys.
func (g *Garlic) DialContext(ctx context.Context, net, addr string) (net.Conn, error) {
	log.WithFields(logrus.Fields{
		"network": net,
		"address": addr,
	}).Debug("Attempting to dial with context")
	if !strings.Contains(addr, ".i2p") {
		log.Debug("Non-I2P address detected, returning null connection")
		return &NullConn{}, nil
	}
	var err error
	if g.SAM, err = g.samSession(); err != nil {
		log.WithError(err).Error("Failed to create SAM session")
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	if g.StreamSession, err = g.setupStreamSession(); err != nil {
		log.WithError(err).Error("Failed to setup stream session")
		return nil, fmt.Errorf("onramp Dial: %v", err)
	}
	log.Debug("Attempting to establish connection with context")
	conn, err := g.StreamSession.DialContext(ctx, net, addr)
	if err != nil {
		log.WithError(err).Error("Failed to establish connection")
		return nil, err
	}

	log.Debug("Successfully established connection")
	return conn, nil
	// return g.StreamSession.DialContext(ctx, net, addr)
}

// Close closes the Garlic structure's sessions and listeners.
func (g *Garlic) Close() error {
	log.WithField("name", g.getName()).Debug("Closing Garlic sessions")
	e1 := g.StreamSession.Close()
	var err error
	if e1 != nil {
		log.WithError(e1).Error("Failed to close stream session")
		err = fmt.Errorf("onramp Close: %v", e1)
	} else {
		log.Debug("Stream session closed successfully")
	}
	e2 := g.SAM.Close()
	if e2 != nil {
		log.WithError(e2).Error("Failed to close SAM session")
		err = fmt.Errorf("onramp Close: %v %v", e1, e2)
	} else {
		log.Debug("SAM session closed successfully")
	}

	if err == nil {
		log.Debug("All sessions closed successfully")
	}

	return err
}

// Keys returns the I2PKeys for the Garlic structure. If none
// exist, they are created and stored.
func (g *Garlic) Keys() (*i2pkeys.I2PKeys, error) {
	log.WithFields(logrus.Fields{
		"name":    g.getName(),
		"address": g.getAddr(),
	}).Debug("Retrieving I2P keys")

	keys, err := I2PKeys(g.getName(), g.getAddr())
	if err != nil {
		log.WithError(err).Error("Failed to get I2P keys")
		return &i2pkeys.I2PKeys{}, fmt.Errorf("onramp Keys: %v", err)
	}
	log.Debug("Successfully retrieved I2P keys")
	return &keys, nil
}

func (g *Garlic) DeleteKeys() error {
	// return DeleteGarlicKeys(g.getName())
	log.WithField("name", g.getName()).Debug("Attempting to delete Garlic keys")
	err := DeleteGarlicKeys(g.getName())
	if err != nil {
		log.WithError(err).Error("Failed to delete Garlic keys")
	}
	log.Debug("Successfully deleted Garlic keys")
	return err
}

// NewGarlic returns a new Garlic struct. It is immediately ready to use with
// I2P streaming.
func NewGarlic(tunName, samAddr string, options []string) (*Garlic, error) {
	log.WithFields(logrus.Fields{
		"tunnel_name": tunName,
		"sam_address": samAddr,
		"options":     options,
	}).Debug("Creating new Garlic instance")

	g := new(Garlic)
	g.name = tunName
	g.addr = samAddr
	g.opts = options
	var err error
	if g.SAM, err = g.samSession(); err != nil {
		log.WithError(err).Error("Failed to create SAM session")
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	if g.StreamSession, err = g.setupStreamSession(); err != nil {
		log.WithError(err).Error("Failed to setup stream session")
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}

	log.Debug("Successfully created new Garlic instance")
	return g, nil
}

// DeleteGarlicKeys deletes the key file at the given path as determined by
// keystore + tunName.
// This is permanent and irreversible, and will change the onion service
// address.
func DeleteGarlicKeys(tunName string) error {
	log.WithField("tunnel_name", tunName).Debug("Attempting to delete Garlic keys")
	keystore, err := I2PKeystorePath()
	if err != nil {
		log.WithError(err).Error("Failed to get keystore path")
		return fmt.Errorf("onramp DeleteGarlicKeys: discovery error %v", err)
	}
	keyspath := filepath.Join(keystore, tunName+".i2p.private")
	log.WithField("path", keyspath).Debug("Deleting key file")
	if err := os.Remove(keyspath); err != nil {
		log.WithError(err).WithField("path", keyspath).Error("Failed to delete key file")
		return fmt.Errorf("onramp DeleteGarlicKeys: %v", err)
	}
	log.Debug("Successfully deleted Garlic keys")
	return nil
}

// I2PKeys returns the I2PKeys at the keystore directory for the given
// tunnel name. If none exist, they are created and stored.
func I2PKeys(tunName, samAddr string) (i2pkeys.I2PKeys, error) {
	log.WithFields(logrus.Fields{
		"tunnel_name": tunName,
		"sam_address": samAddr,
	}).Debug("Looking up I2P keys")

	keystore, err := I2PKeystorePath()
	if err != nil {
		log.WithError(err).Error("Failed to get keystore path")
		return i2pkeys.I2PKeys{}, fmt.Errorf("onramp I2PKeys: discovery error %v", err)
	}
	keyspath := filepath.Join(keystore, tunName+".i2p.private")
	log.WithField("path", keyspath).Debug("Checking for existing keys")
	info, err := os.Stat(keyspath)
	if info != nil {
		if info.Size() == 0 {
			log.WithField("path", keyspath).Debug("Keystore empty, will regenerate keys")
			log.Println("onramp I2PKeys: keystore empty, re-generating keys")
		} else {
			log.WithField("path", keyspath).Debug("Found existing keystore")
		}
	}
	if err != nil {
		log.WithField("path", keyspath).Debug("Keys not found, generating new keys")
		sam, err := sam3.NewSAM(samAddr)
		if err != nil {
			log.WithError(err).Error("Failed to create SAM connection")
			return i2pkeys.I2PKeys{}, fmt.Errorf("onramp I2PKeys: SAM error %v", err)
		}
		log.Debug("SAM connection established")
		keys, err := sam.NewKeys(tunName)
		if err != nil {
			log.WithError(err).Error("Failed to generate new keys")
			return i2pkeys.I2PKeys{}, fmt.Errorf("onramp I2PKeys: keygen error %v", err)
		}
		log.Debug("New keys generated successfully")
		if err = i2pkeys.StoreKeys(keys, keyspath); err != nil {
			log.WithError(err).WithField("path", keyspath).Error("Failed to store generated keys")
			return i2pkeys.I2PKeys{}, fmt.Errorf("onramp I2PKeys: store error %v", err)
		}
		log.WithField("path", keyspath).Debug("Successfully stored new keys")
		return keys, nil
	} else {
		log.WithField("path", keyspath).Debug("Loading existing keys")
		keys, err := i2pkeys.LoadKeys(keyspath)
		if err != nil {
			log.WithError(err).WithField("path", keyspath).Error("Failed to load existing keys")
			return i2pkeys.I2PKeys{}, fmt.Errorf("onramp I2PKeys: load error %v", err)
		}
		log.Debug("Successfully loaded existing keys")
		return keys, nil
	}
}

var garlics map[string]*Garlic

// CloseAllGarlic closes all garlics managed by the onramp package. It does not
// affect objects instantiated by an app.
func CloseAllGarlic() {
	log.WithField("count", len(garlics)).Debug("Closing all Garlic connections")
	for i, g := range garlics {
		log.WithFields(logrus.Fields{
			"index": i,
			"name":  g.name,
		}).Debug("Closing Garlic connection")

		log.Println("Closing garlic", g.name)
		CloseGarlic(i)
	}
	log.Debug("All Garlic connections closed")
}

// CloseGarlic closes the Garlic at the given index. It does not affect Garlic
// objects instantiated by an app.
func CloseGarlic(tunName string) {
	log.WithField("tunnel_name", tunName).Debug("Attempting to close Garlic connection")
	g, ok := garlics[tunName]
	if ok {
		log.Debug("Found Garlic connection, closing")
		// g.Close()
		err := g.Close()
		if err != nil {
			log.WithError(err).Error("Error closing Garlic connection")
		} else {
			log.Debug("Successfully closed Garlic connection")
		}
	} else {
		log.Debug("No Garlic connection found for tunnel name")
	}
}

// SAM_ADDR is the default I2P SAM address. It can be overridden by the
// struct or by changing this variable.
var SAM_ADDR = "127.0.0.1:7656"

// ListenGarlic returns a net.Listener for a garlic structure's keys
// corresponding to a structure managed by the onramp library
// and not instantiated by an app.
func ListenGarlic(network, keys string) (net.Listener, error) {
	log.WithFields(logrus.Fields{
		"network":  network,
		"keys":     keys,
		"sam_addr": SAM_ADDR,
	}).Debug("Creating new Garlic listener")
	g, err := NewGarlic(keys, SAM_ADDR, OPT_DEFAULTS)
	if err != nil {
		log.WithError(err).Error("Failed to create new Garlic")
		return nil, fmt.Errorf("onramp Listen: %v", err)
	}
	garlics[keys] = g
	log.Debug("Successfully created Garlic listener")
	return g.Listen()
}

// DialGarlic returns a net.Conn for a garlic structure's keys
// corresponding to a structure managed by the onramp library
// and not instantiated by an app.
func DialGarlic(network, addr string) (net.Conn, error) {
	log.WithFields(logrus.Fields{
		"network":  network,
		"address":  addr,
		"sam_addr": SAM_ADDR,
	}).Debug("Creating new Garlic connection")

	g, err := NewGarlic(addr, SAM_ADDR, OPT_DEFAULTS)
	if err != nil {
		log.WithError(err).Error("Failed to create new Garlic")
		return nil, fmt.Errorf("onramp Dial: %v", err)
	}
	garlics[addr] = g
	log.WithField("address", addr).Debug("Attempting to dial")
	conn, err := g.Dial(network, addr)
	if err != nil {
		log.WithError(err).Error("Failed to dial connection")
		return nil, err
	}

	log.Debug("Successfully established Garlic connection")
	return conn, nil
	// return g.Dial(network, addr)
}
