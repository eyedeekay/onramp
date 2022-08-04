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

SAM_ADDR is the default I2P SAM address. It can be overridden by the
struct or by changing this variable.

```golang
var SAM_ADDR = "127.0.0.1:7656"
```

## Functions

### func [CloseAllGarlic](/garlic.go#L211)

`func CloseAllGarlic()`

CloseAllGarlic closes all garlics managed by the onramp package. It does not
affect objects instantiated by an app.

### func [CloseAllOnion](/onion.go#L174)

`func CloseAllOnion()`

CloseAllOnion closes all onions managed by the onramp package. It does not
affect objects instantiated by an app.

### func [CloseGarlic](/garlic.go#L220)

`func CloseGarlic(tunName string)`

CloseGarlic closes the Garlic at the given index. It does not affect Garlic
objects instantiated by an app.

### func [CloseOnion](/onion.go#L183)

`func CloseOnion(tunName string)`

CloseOnion closes the Onion at the given index. It does not affect Onion
objects instantiated by an app.

### func [DeleteGarlicKeys](/garlic.go#L159)

`func DeleteGarlicKeys(tunName string) error`

DeleteGarlicKeys deletes the key file at the given path as determined by
keystore + tunName.
This is permanent and irreversible, and will change the onion service
address.

### func [DeleteI2PKeyStore](/common.go#L57)

`func DeleteI2PKeyStore() error`

DeleteI2PKeyStore deletes the I2P Keystore.

### func [DeleteOnionKeys](/onion.go#L216)

`func DeleteOnionKeys(tunName string) error`

DeleteOnionKeys deletes the key file at the given path as determined by
keystore + tunName.

### func [DeleteTorKeyStore](/common.go#L75)

`func DeleteTorKeyStore() error`

DeleteTorKeyStore deletes the Onion Keystore.

### func [DialGarlic](/garlic.go#L246)

`func DialGarlic(network, addr string) (net.Conn, error)`

DialGarlic returns a net.Conn for a garlic structure's keys
corresponding to a structure managed by the onramp library
and not instantiated by an app.

### func [DialOnion](/onion.go#L205)

`func DialOnion(network, addr string) (net.Conn, error)`

DialOnion returns a net.Conn for a onion structure's keys
corresponding to a structure managed by the onramp library
and not instantiated by an app.

### func [GetJoinedWD](/common.go#L14)

`func GetJoinedWD(dir string) (string, error)`

GetJoinedWD returns the working directory joined with the given path.

### func [I2PKeys](/garlic.go#L173)

`func I2PKeys(tunName, samAddr string) (i2pkeys.I2PKeys, error)`

I2PKeys returns the I2PKeys at the keystore directory for the given
tunnel name. If none exist, they are created and stored.

### func [I2PKeystorePath](/common.go#L46)

`func I2PKeystorePath() (string, error)`

I2PKeystorePath returns the path to the I2P Keystore. If the
path is not set, it returns the default path. If the path does
not exist, it creates it.

### func [ListenGarlic](/garlic.go#L234)

`func ListenGarlic(network, keys string) (net.Listener, error)`

ListenGarlic returns a net.Listener for a garlic structure's keys
corresponding to a structure managed by the onramp library
and not instantiated by an app.

### func [ListenOnion](/onion.go#L193)

`func ListenOnion(network, keys string) (net.Listener, error)`

ListenOnion returns a net.Listener for a onion structure's keys
corresponding to a structure managed by the onramp library
and not instantiated by an app.

### func [TorKeys](/onion.go#L135)

`func TorKeys(keyName string) (ed25519.KeyPair, error)`

TorKeys returns a key pair which will be stored at the given key
name in the key store. If the key already exists, it will be
returned. If it does not exist, it will be generated.

### func [TorKeystorePath](/common.go#L64)

`func TorKeystorePath() (string, error)`

TorKeystorePath returns the path to the Onion Keystore. If the
path is not set, it returns the default path. If the path does
not exist, it creates it.

## Types

### type [Garlic](/garlic.go#L19)

`type Garlic struct { ... }`

Garlic is a ready-made I2P streaming manager. Once initialized it always
has a valid I2PKeys and StreamSession.

#### func [NewGarlic](/garlic.go#L140)

`func NewGarlic(tunName, samAddr string, options []string) (*Garlic, error)`

NewGarlic returns a new Garlic struct. It is immediately ready to use with
I2P streaming.

#### func (*Garlic) [Close](/garlic.go#L111)

`func (g *Garlic) Close() error`

Close closes the Garlic structure's sessions and listeners.

#### func (*Garlic) [DeleteKeys](/garlic.go#L134)

`func (g *Garlic) DeleteKeys() error`

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

### type [Onion](/onion.go#L24)

`type Onion struct { ... }`

Onion represents a structure which manages an onion service and
a Tor client. The onion service will automatically have persistent
keys.

#### func [NewOnion](/onion.go#L126)

`func NewOnion(name string) (*Onion, error)`

NewOnion returns a new Onion object.

#### func (*Onion) [Close](/onion.go#L109)

`func (o *Onion) Close() error`

Close closes the Onion Service and all associated resources.

#### func (*Onion) [DeleteKeys](/onion.go#L121)

`func (g *Onion) DeleteKeys() error`

DeleteKeys deletes the keys at the given key name in the key store.
This is permanent and irreversible, and will change the onion service
address.

#### func (*Onion) [Dial](/onion.go#L104)

`func (o *Onion) Dial(net, addr string) (net.Conn, error)`

Dial returns a net.Conn to the given onion address or clearnet address.

#### func (*Onion) [Keys](/onion.go#L114)

`func (o *Onion) Keys() (ed25519.KeyPair, error)`

Keys returns the keys for the Onion

#### func (*Onion) [Listen](/onion.go#L99)

`func (o *Onion) Listen() (net.Listener, error)`

ListenOnion returns a net.Listener which will listen on an onion
address, and will automatically generate a keypair and store it.

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
