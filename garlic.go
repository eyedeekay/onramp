package onramp

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/eyedeekay/i2pkeys"
	"github.com/eyedeekay/sam3"
)

// Garlic is a ready-made I2P streaming manager. Once initialized it always
// has a valid I2PKeys and StreamSession.
type Garlic struct {
	*sam3.StreamListener
	*sam3.StreamSession
	i2pkeys.I2PKeys
	*sam3.SAM
	name string
	addr string
	opts []string
}

// NewGarlic returns a new Garlic struct. It is immediately ready to use with
// I2P streaming.
func NewGarlic(tunName, samAddr string, options []string) (*Garlic, error) {
	g := new(Garlic)
	var err error
	g.name = tunName
	g.addr = samAddr
	g.opts = options
	if g.SAM, err = g.samSession(); err != nil {
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	if g.StreamSession, err = g.setupStreamSession(); err != nil {
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	return g, nil
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

func (g *Garlic) getName() string {
	if g.name == "" {
		return "onramp"
	}
	return g.name
}

func (g *Garlic) getAddr() string {
	if g.addr == "" {
		return "http://localhost:7656"
	}
	return g.addr
}

func (g *Garlic) getOptions() []string {
	if g.opts == nil {
		return sam3.Options_Default
	}
	return g.opts
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

// Listen returns a net.Listener for the Garlic structure's I2P keys.
func (g *Garlic) Listen() (net.Listener, error) {
	var err error
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

// Dial returns a net.Conn for the Garlic structure's I2P keys.
func (g *Garlic) Dial(net, addr string) (net.Conn, error) {
	var err error
	if g.StreamSession, err = g.setupStreamSession(); err != nil {
		return nil, fmt.Errorf("onramp Dial: %v", err)
	}
	return g.StreamSession.Dial(net, addr)
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
