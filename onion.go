//go:build !gen
// +build !gen

package onramp

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
)

var torp *tor.Tor

// Onion represents a structure which manages an onion service and
// a Tor client. The onion service will automatically have persistent
// keys.
type Onion struct {
	*tor.StartConf
	*tor.ListenConf
	*tor.DialConf
	context.Context
	name string
}

func (o *Onion) getStartConf() *tor.StartConf {
	if o.StartConf == nil {
		o.StartConf = &tor.StartConf{}
	}
	return o.StartConf
}

func (o *Onion) getContext() context.Context {
	if o.Context == nil {
		o.Context = context.Background()
	}
	return o.Context
}

func (o *Onion) getListenConf() *tor.ListenConf {
	keys, err := o.Keys()
	if err != nil {
		log.Fatalf("Unable to get onion service keys, %s", err)
	}
	if o.ListenConf == nil {
		o.ListenConf = &tor.ListenConf{
			Key: keys,
		}
	}
	return o.ListenConf
}

func (o *Onion) getDialConf() *tor.DialConf {
	if o.DialConf == nil {
		o.DialConf = &tor.DialConf{}
	}
	return o.DialConf
}

func (o *Onion) getTor() *tor.Tor {
	if torp == nil {
		var err error
		torp, err = tor.Start(o.getContext(), o.getStartConf())
		if err != nil {
			panic(err)
		}
	}
	return torp
}

func (o *Onion) getDialer() *tor.Dialer {
	//if o.Dialer == nil {
	//var err error
	//o.Dialer, err
	dialer, err := o.getTor().Dialer(o.getContext(), o.getDialConf())
	if err != nil {
		panic(err)
	}
	//}
	//return o.Dialer
	return dialer
}

func (o *Onion) getName() string {
	if o.name == "" {
		o.name = "onramp-onion"
	}
	return o.name
}

// NewListener returns a net.Listener which will listen on an onion
// address, and will automatically generate a keypair and store it.
// the args are always ignored
func (o *Onion) NewListener(n, addr string) (net.Listener, error) {
	return o.Listen(n)
}

// Listen returns a net.Listener which will listen on an onion
// address, and will automatically generate a keypair and store it.
// the args are always ignored
func (o *Onion) Listen(args ...string) (net.Listener, error) {
	return o.OldListen(args...)
}

// OldListen returns a net.Listener which will listen on an onion
// address, and will automatically generate a keypair and store it.
// the args are always ignored
func (o *Onion) OldListen(args ...string) (net.Listener, error) {
	return o.getTor().Listen(o.getContext(), o.getListenConf())
}

// ListenTLS returns a net.Listener which will apply TLS encryption
// to the onion listener, which will not be decrypted until it reaches
// the browser
func (o *Onion) ListenTLS(args ...string) (net.Listener, error) {
	cert, err := o.TLSKeys()
	if err != nil {
		return nil, fmt.Errorf("onramp ListenTLS: %v", err)
	}
	l, err := o.getTor().Listen(o.getContext(), o.getListenConf())
	return tls.NewListener(
		l,
		&tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	), nil
}

// Dial returns a net.Conn to the given onion address or clearnet address.
func (o *Onion) Dial(net, addr string) (net.Conn, error) {
	return o.getDialer().DialContext(o.getContext(), net, addr)
}

// Close closes the Onion Service and all associated resources.
func (o *Onion) Close() error {
	return o.getTor().Close()
}

// Keys returns the keys for the Onion
func (o *Onion) Keys() (ed25519.KeyPair, error) {
	return TorKeys(o.getName())
}

// DeleteKeys deletes the keys at the given key name in the key store.
// This is permanent and irreversible, and will change the onion service
// address.
func (g *Onion) DeleteKeys() error {
	return DeleteOnionKeys(g.getName())
}

// NewOnion returns a new Onion object.
func NewOnion(name string) (*Onion, error) {
	return &Onion{
		name: name,
	}, nil
}

// TorKeys returns a key pair which will be stored at the given key
// name in the key store. If the key already exists, it will be
// returned. If it does not exist, it will be generated.
func TorKeys(keyName string) (ed25519.KeyPair, error) {
	keystore, err := TorKeystorePath()
	if err != nil {
		return nil, fmt.Errorf("onramp OnionKeys: discovery error %v", err)
	}
	var keys ed25519.KeyPair
	keysPath := filepath.Join(keystore, keyName+".tor.private")
	if _, err := os.Stat(keysPath); os.IsNotExist(err) {
		tkeys, err := ed25519.GenerateKey(nil)
		if err != nil {
			log.Fatalf("Unable to generate onion service key, %s", err)
		}
		keys = tkeys
		f, err := os.Create(keysPath)
		if err != nil {
			log.Fatalf("Unable to create Tor keys file for writing, %s", err)
		}
		defer f.Close()
		_, err = f.Write(tkeys.PrivateKey())
		if err != nil {
			log.Fatalf("Unable to write Tor keys to disk, %s", err)
		}
	} else if err == nil {
		tkeys, err := ioutil.ReadFile(keysPath)
		if err != nil {
			log.Fatalf("Unable to read Tor keys from disk")
		}
		k := ed25519.FromCryptoPrivateKey(tkeys)
		keys = k
	} else {
		log.Fatalf("Unable to set up Tor keys, %s", err)
	}
	return keys, nil
}

var onions map[string]*Onion

// CloseAllOnion closes all onions managed by the onramp package. It does not
// affect objects instantiated by an app.
func CloseAllOnion() {
	for i, g := range onions {
		log.Println("Closing onion", g.name)
		CloseOnion(i)
	}
}

// CloseOnion closes the Onion at the given index. It does not affect Onion
// objects instantiated by an app.
func CloseOnion(tunName string) {
	g, ok := onions[tunName]
	if ok {
		g.Close()
	}
}

// ListenOnion returns a net.Listener for a onion structure's keys
// corresponding to a structure managed by the onramp library
// and not instantiated by an app.
func ListenOnion(network, keys string) (net.Listener, error) {
	g, err := NewOnion(keys)
	if err != nil {
		return nil, fmt.Errorf("onramp Listen: %v", err)
	}
	onions[keys] = g
	return g.Listen()
}

// DialOnion returns a net.Conn for a onion structure's keys
// corresponding to a structure managed by the onramp library
// and not instantiated by an app.
func DialOnion(network, addr string) (net.Conn, error) {
	g, err := NewOnion(addr)
	if err != nil {
		return nil, fmt.Errorf("onramp Dial: %v", err)
	}
	onions[addr] = g
	return g.Dial(network, addr)
}

// DeleteOnionKeys deletes the key file at the given path as determined by
// keystore + tunName.
func DeleteOnionKeys(tunName string) error {
	keystore, err := TorKeystorePath()
	if err != nil {
		return fmt.Errorf("onramp DeleteOnionKeys: discovery error %v", err)
	}
	keyspath := filepath.Join(keystore, tunName+".i2p.private")
	if err := os.Remove(keyspath); err != nil {
		return fmt.Errorf("onramp DeleteOnionKeys: %v", err)
	}
	return nil
}
