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

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
