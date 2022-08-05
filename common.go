//go:build !gen
// +build !gen

package onramp

import (
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

//go:generate go run -tags gen ./gen.go

// GetJoinedWD returns the working directory joined with the given path.
func GetJoinedWD(dir string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	jwd := filepath.Join(wd, dir)
	ajwd, err := filepath.Abs(jwd)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(ajwd); err != nil {
		os.MkdirAll(ajwd, 0755)
	}
	return ajwd, nil
}

var i2pdefault, i2pkserr = GetJoinedWD("i2pkeys")
var tordefault, torkserr = GetJoinedWD("onionkeys")
var tlsdefault, tlskserr = GetJoinedWD("tlskeys")

// I2P_KEYSTORE_PATH is the place where I2P Keys will be saved.
// it defaults to the directory "i2pkeys" current working directory.
// reference it by calling I2PKeystorePath() to check for errors
var I2P_KEYSTORE_PATH = i2pdefault

// ONION_KEYSTORE_PATH is the place where Onion Keys will be saved.
// it defaults to the directory "onionkeys" current working directory.
// reference it by calling OnionKeystorePath() to check for errors
var ONION_KEYSTORE_PATH = tordefault

// TLS_KEYSTORE_PATH is the place where TLS Keys will be saved.
// it defaults to the directory "tlskeys" current working directory.
// reference it by calling TLSKeystorePath() to check for errors
var TLS_KEYSTORE_PATH = tlsdefault

// I2PKeystorePath returns the path to the I2P Keystore. If the
// path is not set, it returns the default path. If the path does
// not exist, it creates it.
func I2PKeystorePath() (string, error) {
	if _, err := os.Stat(I2P_KEYSTORE_PATH); err != nil {
		err := os.MkdirAll(I2P_KEYSTORE_PATH, 0755)
		if err != nil {
			return "", err
		}
	}
	return I2P_KEYSTORE_PATH, nil
}

// DeleteI2PKeyStore deletes the I2P Keystore.
func DeleteI2PKeyStore() error {
	return os.RemoveAll(I2P_KEYSTORE_PATH)
}

// TorKeystorePath returns the path to the Onion Keystore. If the
// path is not set, it returns the default path. If the path does
// not exist, it creates it.
func TorKeystorePath() (string, error) {
	if _, err := os.Stat(ONION_KEYSTORE_PATH); err != nil {
		err := os.MkdirAll(ONION_KEYSTORE_PATH, 0755)
		if err != nil {
			return "", err
		}
	}
	return ONION_KEYSTORE_PATH, nil
}

// DeleteTorKeyStore deletes the Onion Keystore.
func DeleteTorKeyStore() error {
	return os.RemoveAll(ONION_KEYSTORE_PATH)
}

// TLSKeystorePath returns the path to the TLS Keystore. If the
// path is not set, it returns the default path. If the path does
// not exist, it creates it.
func TLSKeystorePath() (string, error) {
	if _, err := os.Stat(TLS_KEYSTORE_PATH); err != nil {
		err := os.MkdirAll(TLS_KEYSTORE_PATH, 0755)
		if err != nil {
			return "", err
		}
	}
	return TLS_KEYSTORE_PATH, nil
}

// DeleteTLSKeyStore deletes the TLS Keystore.
func DeleteTLSKeyStore() error {
	return os.RemoveAll(TLS_KEYSTORE_PATH)
}

// Dial returns a connection for the given network and address.
// network is ignored. If the address ends in i2p, it returns an I2P connection.
// if the address ends in anything else, it returns a Tor connection.
func Dial(network, addr string) (net.Conn, error) {
	url, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	hostname := url.Hostname()
	if strings.HasSuffix(hostname, ".i2p") {
		return DialGarlic(network, addr)
	}
	return DialOnion(network, addr)
}

// Listen returns a listener for the given network and address.
// if network is i2p or garlic, it returns an I2P listener.
// if network is tor or onion, it returns an Onion listener.
// if keys ends with ".i2p", it returns an I2P listener.
func Listen(network, keys string) (net.Listener, error) {
	if network == "i2p" || network == "garlic" {
		return ListenGarlic(network, keys)
	}
	if network == "tor" || network == "onion" {
		return ListenOnion(network, keys)
	}
	url, err := url.Parse(keys)
	if err != nil {
		return nil, err
	}
	hostname := url.Hostname()
	if strings.HasSuffix(hostname, ".i2p") {
		return ListenGarlic(network, keys)
	}
	return ListenOnion(network, keys)
}
