//go:build !gen
// +build !gen

package onramp

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/eyedeekay/i2pkeys"
	"github.com/eyedeekay/sam3"
)

// Garlic is a ready-made I2P streaming manager. Once initialized it always
// has a valid I2PKeys and StreamSession.
type Garlic struct {
	*sam3.StreamListener
	*sam3.StreamSession
	*sam3.DatagramSession
	i2pkeys.I2PKeys
	*sam3.SAM
	name string
	addr string
	opts []string
}

var OPT_DEFAULTS = sam3.Options_Default

func (g *Garlic) getName() string {
	if g.name == "" {
		return "onramp"
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
		var err error
		g.SAM, err = sam3.NewSAM(g.getAddr())
		if err != nil {
			return nil, fmt.Errorf("onramp samSession: %v", err)
		}
	}
	return g.SAM, nil
}

func (g Garlic) setupStreamSession() (*sam3.StreamSession, error) {
	if g.StreamSession == nil {
		var err error
		g.I2PKeys, err = g.Keys()
		if err != nil {
			return nil, fmt.Errorf("onramp setupStreamSession: %v", err)
		}
		log.Println("Creating stream session with keys:", g.I2PKeys.Address.Base32())
		g.StreamSession, err = g.SAM.NewStreamSession(g.getName(), g.I2PKeys, g.getOptions())
		if err != nil {
			return nil, fmt.Errorf("onramp setupStreamSession: %v", err)
		}
		return g.StreamSession, nil
	}
	return g.StreamSession, nil
}

func (g Garlic) setupDatagramSession() (*sam3.DatagramSession, error) {
	if g.DatagramSession == nil {
		var err error
		g.I2PKeys, err = g.Keys()
		if err != nil {
			return nil, fmt.Errorf("onramp setupDatagramSession: %v", err)
		}
		log.Println("Creating datagram session with keys:", g.I2PKeys.Address.Base32())
		g.DatagramSession, err = g.SAM.NewDatagramSession(g.getName(), g.I2PKeys, g.getOptions(), 0)
		if err != nil {
			return nil, fmt.Errorf("onramp setupDatagramSession: %v", err)
		}
		return g.DatagramSession, nil
	}
	return g.DatagramSession, nil
}

// Listen returns a net.Listener for the Garlic structure's I2P keys.
func (g *Garlic) Listen() (net.Listener, error) {
	var err error
	if g.SAM, err = g.samSession(); err != nil {
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	if g.StreamSession, err = g.setupStreamSession(); err != nil {
		return nil, fmt.Errorf("onramp Listen: %v", err)
	}
	if g.StreamListener == nil {
		g.StreamListener, err = g.StreamSession.Listen()
		if err != nil {
			return nil, fmt.Errorf("onramp Listen: %v", err)
		}
	}
	return g.StreamListener, nil
}

// ListenPacket returns a net.PacketConn for the Garlic structure's I2P keys.
func (g *Garlic) ListenPacket() (net.PacketConn, error) {
	var err error
	if g.SAM, err = g.samSession(); err != nil {
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	if g.DatagramSession, err = g.setupDatagramSession(); err != nil {
		return nil, fmt.Errorf("onramp Listen: %v", err)
	}
	return g.DatagramSession, nil
}

// ListenTLS returns a net.Listener for the Garlic structure's I2P keys,
// which also uses TLS either for additional encryption, authentication,
// or browser-compatibility.
func (g *Garlic) ListenTLS() (net.Listener, error) {
	var err error
	if g.SAM, err = g.samSession(); err != nil {
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	if g.StreamSession, err = g.setupStreamSession(); err != nil {
		return nil, fmt.Errorf("onramp Listen: %v", err)
	}
	if g.StreamListener == nil {
		g.StreamListener, err = g.StreamSession.Listen()
		if err != nil {
			return nil, fmt.Errorf("onramp Listen: %v", err)
		}
	}
	cert, err := g.TLSKeys()
	if err != nil {
		return nil, fmt.Errorf("onramp ListenTLS: %v", err)
	}
	return tls.NewListener(
		g.StreamListener,
		&tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	), nil
}

// Dial returns a net.Conn for the Garlic structure's I2P keys.
func (g *Garlic) Dial(net, addr string) (net.Conn, error) {
	if !strings.Contains(addr, ".i2p") {
		return &NullConn{}, nil
	}
	var err error
	if g.SAM, err = g.samSession(); err != nil {
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	if g.StreamSession, err = g.setupStreamSession(); err != nil {
		return nil, fmt.Errorf("onramp Dial: %v", err)
	}
	return g.StreamSession.Dial(net, addr)
}

// DialContext returns a net.Conn for the Garlic structure's I2P keys.
func (g *Garlic) DialContext(ctx context.Context, net, addr string) (net.Conn, error) {
	if !strings.Contains(addr, ".i2p") {
		return &NullConn{}, nil
	}
	var err error
	if g.SAM, err = g.samSession(); err != nil {
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	if g.StreamSession, err = g.setupStreamSession(); err != nil {
		return nil, fmt.Errorf("onramp Dial: %v", err)
	}
	return g.StreamSession.DialContext(ctx, net, addr)
}

// Close closes the Garlic structure's sessions and listeners.
func (g *Garlic) Close() error {
	e1 := g.StreamSession.Close()
	var err error
	if e1 != nil {
		err = fmt.Errorf("onramp Close: %v", e1)
	}
	e2 := g.SAM.Close()
	if e2 != nil {
		err = fmt.Errorf("onramp Close: %v %v", e1, e2)
	}
	return err
}

// Keys returns the I2PKeys for the Garlic structure. If none
// exist, they are created and stored.
func (g *Garlic) Keys() (i2pkeys.I2PKeys, error) {
	keys, err := I2PKeys(g.getName(), g.getAddr())
	if err != nil {
		return i2pkeys.I2PKeys{}, fmt.Errorf("onramp Keys: %v", err)
	}
	return keys, nil
}

func (g *Garlic) DeleteKeys() error {
	return DeleteGarlicKeys(g.getName())
}

// NewGarlic returns a new Garlic struct. It is immediately ready to use with
// I2P streaming.
func NewGarlic(tunName, samAddr string, options []string) (*Garlic, error) {
	g := new(Garlic)
	g.name = tunName
	g.addr = samAddr
	g.opts = options
	var err error
	if g.SAM, err = g.samSession(); err != nil {
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	if g.StreamSession, err = g.setupStreamSession(); err != nil {
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	return g, nil
}

// DeleteGarlicKeys deletes the key file at the given path as determined by
// keystore + tunName.
// This is permanent and irreversible, and will change the onion service
// address.
func DeleteGarlicKeys(tunName string) error {
	keystore, err := I2PKeystorePath()
	if err != nil {
		return fmt.Errorf("onramp DeleteGarlicKeys: discovery error %v", err)
	}
	keyspath := filepath.Join(keystore, tunName+".i2p.private")
	if err := os.Remove(keyspath); err != nil {
		return fmt.Errorf("onramp DeleteGarlicKeys: %v", err)
	}
	return nil
}

// I2PKeys returns the I2PKeys at the keystore directory for the given
// tunnel name. If none exist, they are created and stored.
func I2PKeys(tunName, samAddr string) (i2pkeys.I2PKeys, error) {
	keystore, err := I2PKeystorePath()
	if err != nil {
		return i2pkeys.I2PKeys{}, fmt.Errorf("onramp I2PKeys: discovery error %v", err)
	}
	keyspath := filepath.Join(keystore, tunName+".i2p.private")
	info, err := os.Stat(keyspath)
	if info != nil {
		if info.Size() == 0 {
			log.Println("onramp I2PKeys: keystore empty, re-generating keys")
		}
	}
	if err != nil {
		sam, err := sam3.NewSAM(samAddr)
		if err != nil {
			return i2pkeys.I2PKeys{}, fmt.Errorf("onramp I2PKeys: SAM error %v", err)
		}
		keys, err := sam.NewKeys(tunName)
		if err != nil {
			return i2pkeys.I2PKeys{}, fmt.Errorf("onramp I2PKeys: keygen error %v", err)
		}
		if err = i2pkeys.StoreKeys(keys, keyspath); err != nil {
			return i2pkeys.I2PKeys{}, fmt.Errorf("onramp I2PKeys: store error %v", err)
		}
		return keys, nil
	} else {
		keys, err := i2pkeys.LoadKeys(keyspath)
		if err != nil {
			return i2pkeys.I2PKeys{}, fmt.Errorf("onramp I2PKeys: load error %v", err)
		}
		return keys, nil
	}
}

var garlics map[string]*Garlic

// CloseAllGarlic closes all garlics managed by the onramp package. It does not
// affect objects instantiated by an app.
func CloseAllGarlic() {
	for i, g := range garlics {
		log.Println("Closing garlic", g.name)
		CloseGarlic(i)
	}
}

// CloseGarlic closes the Garlic at the given index. It does not affect Garlic
// objects instantiated by an app.
func CloseGarlic(tunName string) {
	g, ok := garlics[tunName]
	if ok {
		g.Close()
	}
}

// SAM_ADDR is the default I2P SAM address. It can be overridden by the
// struct or by changing this variable.
var SAM_ADDR = "127.0.0.1:7656"

// ListenGarlic returns a net.Listener for a garlic structure's keys
// corresponding to a structure managed by the onramp library
// and not instantiated by an app.
func ListenGarlic(network, keys string) (net.Listener, error) {
	g, err := NewGarlic(keys, SAM_ADDR, OPT_DEFAULTS)
	if err != nil {
		return nil, fmt.Errorf("onramp Listen: %v", err)
	}
	garlics[keys] = g
	return g.Listen()
}

// DialGarlic returns a net.Conn for a garlic structure's keys
// corresponding to a structure managed by the onramp library
// and not instantiated by an app.
func DialGarlic(network, addr string) (net.Conn, error) {
	g, err := NewGarlic(addr, SAM_ADDR, OPT_DEFAULTS)
	if err != nil {
		return nil, fmt.Errorf("onramp Dial: %v", err)
	}
	garlics[addr] = g
	return g.Dial(network, addr)
}
