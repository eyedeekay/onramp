# Onramp I2P and Tor Library

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/.)
[![Go Report Card](https://goreportcard.com/badge/.)](https://goreportcard.com/report/.)

## Variables

I2P_KEYSTORE_PATH is the place where I2P Keys will be saved.
it defaults to the directory "i2pkeys" current working directory
reference it by calling I2PKeystorePath() to check for errors

```golang
var I2P_KEYSTORE_PATH = i2pdefault
```

ONION_KEYSTORE_PATH is the place where Onion Keys will be saved.
it defaults to the directory "onionkeys" current working directory
reference it by calling OnionKeystorePath() to check for errors

```golang
var ONION_KEYSTORE_PATH = tordefault
```

```golang
var OPT_DEFAULTS = sam3.Options_Default
```

## Functions

### func [Close](/garlic.go#L200)

`func Close(tunName string)`

Close closes the Garlic at the given index. It does not affect Garlic
objects instantiated by an app.

### func [CloseAll](/garlic.go#L191)

`func CloseAll()`

Close() closes all garlics managed by the onramp package. It does not
affect objects instantiated by an app.

### func [Dial](/garlic.go#L222)

`func Dial(network, addr string) (net.Conn, error)`

Dial returns a net.Conn for a garlic structure's keys
corresponding to a structure managed by the onramp library
and not instantiated by an app.

### func [GetJoinedWD](/common.go#L14)

`func GetJoinedWD(dir string) (string, error)`

GetJoinedWD returns the working directory joined with the given path.

### func [I2PKeys](/garlic.go#L153)

`func I2PKeys(tunName, samAddr string) (i2pkeys.I2PKeys, error)`

I2PKeys returns the I2PKeys at the keystore directory for the given
tunnel name. If none exist, they are created and stored.

### func [I2PKeystorePath](/common.go#L46)

`func I2PKeystorePath() (string, error)`

I2PKeystorePath returns the path to the I2P Keystore. If the
path is not set, it returns the default path. If the path does
not exist, it creates it.

### func [Listen](/garlic.go#L210)

`func Listen(network, keys string) (net.Listener, error)`

Listen returns a net.Listener for a garlic structure's keys
corresponding to a structure managed by the onramp library
and not instantiated by an app.

### func [TorKeystorePath](/common.go#L59)

`func TorKeystorePath() (string, error)`

TorKeystorePath returns the path to the Onion Keystore. If the
path is not set, it returns the default path. If the path does
not exist, it creates it.

## Types

### type [Garlic](/garlic.go#L19)

`type Garlic struct { ... }`

Garlic is a ready-made I2P streaming manager. Once initialized it always
has a valid I2PKeys and StreamSession.

#### func [NewGarlic](/garlic.go#L136)

`func NewGarlic(tunName, samAddr string, options []string) (*Garlic, error)`

NewGarlic returns a new Garlic struct. It is immediately ready to use with
I2P streaming.

#### func (*Garlic) [Close](/garlic.go#L111)

`func (g *Garlic) Close() error`

Close closes the Garlic structure's sessions and listeners.

#### func (*Garlic) [Dial](/garlic.go#L99)

`func (g *Garlic) Dial(net, addr string) (net.Conn, error)`

Dial returns a net.Conn for the Garlic structure's I2P keys.

#### func (*Garlic) [Keys](/garlic.go#L126)

`func (g *Garlic) Keys() (i2pkeys.I2PKeys, error)`

Keys returns the I2PKeys for the Garlic structure. If none
exist, they are created and stored.

#### func (*Garlic) [Listen](/garlic.go#L81)

`func (g *Garlic) Listen() (net.Listener, error)`

Listen returns a net.Listener for the Garlic structure's I2P keys.

### type [Onion](/onion.go#L6)

`type Onion struct { ... }`

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
