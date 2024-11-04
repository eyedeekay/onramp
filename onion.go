//go:build !gen
// +build !gen

package onramp

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/sirupsen/logrus"
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
		log.Debug("Initializing new Tor instance")
		var err error
		torp, err = tor.Start(o.getContext(), o.getStartConf())
		if err != nil {
			log.WithError(err).Error("Failed to start Tor")
			panic(err) //return nil instead?
		}
		log.Debug("Tor instance started successfully")
	}
	return torp
}

func (o *Onion) getDialer() *tor.Dialer {
	//if o.Dialer == nil {
	//var err error
	//o.Dialer, err
	log.Debug("Creating new Tor dialer")
	dialer, err := o.getTor().Dialer(o.getContext(), o.getDialConf())
	if err != nil {
		log.WithError(err).Error("Failed to create Tor dialer")
		panic(err)
	}
	log.Debug("Tor dialer created successfully")
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
	log.WithFields(logrus.Fields{
		"args": args,
		"name": o.getName(),
	}).Debug("Setting up Onion listener")
	listener, err := o.OldListen(args...)
	if err != nil {
		log.WithError(err).Error("Failed to create Onion listener")
		return nil, err
	}

	log.Debug("Successfully created Onion listener")
	return listener, nil
	//return o.OldListen(args...)
}

// OldListen returns a net.Listener which will listen on an onion
// address, and will automatically generate a keypair and store it.
// the args are always ignored
func (o *Onion) OldListen(args ...string) (net.Listener, error) {
	log.WithField("name", o.getName()).Debug("Creating Tor listener")

	listener, err := o.getTor().Listen(o.getContext(), o.getListenConf())
	if err != nil {
		log.WithError(err).Error("Failed to create Tor listener")
		return nil, err
	}

	log.Debug("Successfully created Tor listener")
	return listener, nil
	//return o.getTor().Listen(o.getContext(), o.getListenConf())
}

// ListenTLS returns a net.Listener which will apply TLS encryption
// to the onion listener, which will not be decrypted until it reaches
// the browser
func (o *Onion) ListenTLS(args ...string) (net.Listener, error) {
	log.WithField("args", args).Debug("Setting up TLS Onion listener")
	cert, err := o.TLSKeys()
	if err != nil {
		log.WithError(err).Error("Failed to get TLS keys")
		return nil, fmt.Errorf("onramp ListenTLS: %v", err)
	}
	log.Debug("Creating base Tor listener")
	l, err := o.getTor().Listen(o.getContext(), o.getListenConf())
	if err != nil {
		log.WithError(err).Error("Failed to create base Tor listener")
		return nil, err
	}
	log.Debug("Wrapping Tor listener with TLS")
	return tls.NewListener(
		l,
		&tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	), nil
}

// Dial returns a net.Conn to the given onion address or clearnet address.
func (o *Onion) Dial(net, addr string) (net.Conn, error) {
	log.WithFields(logrus.Fields{
		"network": net,
		"address": addr,
	}).Debug("Attempting to dial via Tor")
	conn, err := o.getDialer().DialContext(o.getContext(), net, addr)
	if err != nil {
		log.WithError(err).Error("Failed to establish Tor connection")
		return nil, err
	}

	log.Debug("Successfully established Tor connection")
	return conn, nil
	//return o.getDialer().DialContext(o.getContext(), net, addr)
}

// Close closes the Onion Service and all associated resources.
func (o *Onion) Close() error {
	log.WithField("name", o.getName()).Debug("Closing Onion service")

	err := o.getTor().Close()
	if err != nil {
		log.WithError(err).Error("Failed to close Tor instance")
		return err
	}

	log.Debug("Successfully closed Onion service")
	return nil
	//return o.getTor().Close()
}

// Keys returns the keys for the Onion
func (o *Onion) Keys() (ed25519.KeyPair, error) {
	log.WithField("name", o.getName()).Debug("Retrieving Onion keys")

	keys, err := TorKeys(o.getName())
	if err != nil {
		log.WithError(err).Error("Failed to get Tor keys")
		return nil, err
	}

	log.Debug("Successfully retrieved Onion keys")
	return keys, nil
	//return TorKeys(o.getName())
}

// DeleteKeys deletes the keys at the given key name in the key store.
// This is permanent and irreversible, and will change the onion service
// address.
func (g *Onion) DeleteKeys() error {
	log.WithField("Onion keys", g.getName()).Debug("Deleting Onion keys")
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
	log.WithField("key_name", keyName).Debug("Getting Tor keys")
	keystore, err := TorKeystorePath()
	if err != nil {
		log.WithError(err).Error("Failed to get keystore path")
		return nil, fmt.Errorf("onramp OnionKeys: discovery error %v", err)
	}
	var keys ed25519.KeyPair
	keysPath := filepath.Join(keystore, keyName+".tor.private")
	log.WithField("path", keysPath).Debug("Checking for existing keys")
	if _, err := os.Stat(keysPath); os.IsNotExist(err) {
		log.Debug("Generating new Tor keys")
		tkeys, err := ed25519.GenerateKey(nil)
		if err != nil {
			log.WithError(err).Error("Failed to generate onion service key")
			log.Fatal("Unable to generate onion service key")
		}
		keys = tkeys

		log.WithField("path", keysPath).Debug("Creating key file")
		f, err := os.Create(keysPath)
		if err != nil {
			log.WithError(err).Error("Failed to create Tor keys file")
			log.Fatal("Unable to create Tor keys file for writing")
		}
		defer f.Close()
		_, err = f.Write(tkeys.PrivateKey())
		if err != nil {
			log.WithError(err).Error("Failed to write Tor keys to disk")
			log.Fatal("Unable to write Tor keys to disk")
		}
		log.Debug("Successfully generated and stored new keys")
	} else if err == nil {
		log.Debug("Loading existing Tor keys")
		tkeys, err := os.ReadFile(keysPath)
		if err != nil {
			log.WithError(err).Error("Failed to read Tor keys from disk")
			log.Fatal("Unable to read Tor keys from disk")
		}
		k := ed25519.FromCryptoPrivateKey(tkeys)
		keys = k
		log.Debug("Successfully loaded existing keys")
	} else {
		log.WithError(err).Error("Failed to set up Tor keys")
		log.Fatal("Unable to set up Tor keys")
	}
	return keys, nil
}

var onions map[string]*Onion

// CloseAllOnion closes all onions managed by the onramp package. It does not
// affect objects instantiated by an app.
func CloseAllOnion() {
	log.WithField("count", len(onions)).Debug("Closing all Onion services")
	for i, g := range onions {
		log.WithFields(logrus.Fields{
			"index": i,
			"name":  g.name,
		}).Debug("Closing Onion service")
		CloseOnion(i)
	}

	log.Debug("All Onion services closed")
}

// CloseOnion closes the Onion at the given index. It does not affect Onion
// objects instantiated by an app.
func CloseOnion(tunName string) {
	log.WithField("tunnel_name", tunName).Debug("Attempting to close Onion service")

	g, ok := onions[tunName]
	if ok {
		log.WithField("name", g.name).Debug("Found Onion service, closing")
		err := g.Close()
		if err != nil {
			log.WithError(err).Error("Failed to close Onion service")
		} else {
			log.Debug("Successfully closed Onion service")
		}
	} else {
		log.Debug("No Onion service found for tunnel name")
	}
}

// ListenOnion returns a net.Listener for a onion structure's keys
// corresponding to a structure managed by the onramp library
// and not instantiated by an app.
func ListenOnion(network, keys string) (net.Listener, error) {
	log.WithFields(logrus.Fields{
		"network": network,
		"keys":    keys,
	}).Debug("Creating new Onion listener")

	g, err := NewOnion(keys)
	if err != nil {
		log.WithError(err).Error("Failed to create new Onion")
		return nil, fmt.Errorf("onramp Listen: %v", err)
	}
	onions[keys] = g
	log.Debug("Onion service registered, creating listener")

	listener, err := g.Listen()
	if err != nil {
		log.WithError(err).Error("Failed to create Onion listener")
		return nil, err
	}

	log.Debug("Successfully created Onion listener")
	return listener, nil
	//return g.Listen()
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
	log.WithField("tunnel_name", tunName).Debug("Attempting to delete Onion keys")

	keystore, err := TorKeystorePath()
	if err != nil {
		log.WithError(err).Error("Failed to get keystore path")
		return fmt.Errorf("onramp DeleteOnionKeys: discovery error %v", err)
	}
	keyspath := filepath.Join(keystore, tunName+".i2p.private")
	log.WithError(err).Error("Failed to get keystore path")
	if err := os.Remove(keyspath); err != nil {
		log.WithError(err).WithField("path", keyspath).Error("Failed to delete key file")
		return fmt.Errorf("onramp DeleteOnionKeys: %v", err)
	}
	log.Debug("Successfully deleted Onion keys")
	return nil
}
