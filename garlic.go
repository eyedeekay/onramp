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

type Garlic struct {
	*sam3.StreamListener
	*sam3.StreamSession
	i2pkeys.I2PKeys
	*sam3.SAM
	name string
	addr string
	opts []string
}

func NewGarlic(tunName, samAddr string, options []string) (*Garlic, error) {
	g := new(Garlic)
	var err error
	g.SAM, err = sam3.NewSAM(samAddr)
	if err != nil {
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	g.name = tunName
	g.addr = samAddr
	g.opts = options
	if g.StreamSession, err = g.setupStreamSession(); err != nil {
		return nil, fmt.Errorf("onramp NewGarlic: %v", err)
	}
	return g, nil
}

func (g Garlic) setupStreamSession() (*sam3.StreamSession, error) {
	if g.StreamSession == nil {
		var err error
		g.I2PKeys, err = g.Keys()
		if err != nil {
			return nil, fmt.Errorf("onramp setupStreamSession: %v", err)
		}
		log.Println("Creating stream session with keys:", g.I2PKeys.Address.Base32())
		g.StreamSession, err = g.SAM.NewStreamSession(g.name, g.I2PKeys, g.opts)
		if err != nil {
			return nil, fmt.Errorf("onramp setupStreamSession: %v", err)
		}
		return g.StreamSession, nil
	}
	return g.StreamSession, nil
}

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

func (g *Garlic) Dial(net, addr string) (net.Conn, error) {
	var err error
	if g.StreamSession, err = g.setupStreamSession(); err != nil {
		return nil, fmt.Errorf("onramp Dial: %v", err)
	}
	return g.StreamSession.Dial(net, addr)
}

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

func (g *Garlic) Keys() (i2pkeys.I2PKeys, error) {
	keys, err := I2PKeys(g.name, g.addr)
	if err != nil {
		return i2pkeys.I2PKeys{}, fmt.Errorf("onramp Keys: %v", err)
	}
	return keys, nil
}

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
