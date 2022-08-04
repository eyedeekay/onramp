//go:build !gen
// +build !gen

package onramp

import "github.com/cretz/bine/tor"

type Onion struct {
	*tor.Tor
	*tor.ListenConf
	*tor.Dialer
}

func (o *Onion) getTor() *tor.Tor {
	if o.Tor == nil {
		o.Tor = tor.NewTor()
	}
	return o.Tor
}
