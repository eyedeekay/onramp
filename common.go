//go:build !gen
// +build !gen

package onramp

import (
	"github.com/sirupsen/logrus"
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
		log.WithError(err).Error("Failed to get working directory")
		return "", err
	}
	jwd := filepath.Join(wd, dir)
	ajwd, err := filepath.Abs(jwd)
	if err != nil {
		log.WithError(err).WithField("path", jwd).Error("Failed to get absolute path")
		return "", err
	}
	if _, err := os.Stat(ajwd); err != nil {
		log.WithField("path", ajwd).Debug("Directory does not exist, creating")
		if err := os.MkdirAll(ajwd, 0755); err != nil {
			log.WithError(err).WithField("path", ajwd).Error("Failed to create directory")
			return "", err
		}
	}
	log.WithField("path", ajwd).Debug("Successfully got joined working directory")
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
	log.WithField("path", I2P_KEYSTORE_PATH).Debug("Checking I2P keystore path")
	if _, err := os.Stat(I2P_KEYSTORE_PATH); err != nil {
		log.WithField("path", I2P_KEYSTORE_PATH).Debug("I2P keystore directory does not exist, creating")
		err := os.MkdirAll(I2P_KEYSTORE_PATH, 0755)
		if err != nil {
			log.WithError(err).WithField("path", I2P_KEYSTORE_PATH).Error("Failed to create I2P keystore directory")
			return "", err
		}
	}
	log.WithField("path", I2P_KEYSTORE_PATH).Debug("I2P keystore path verified")
	return I2P_KEYSTORE_PATH, nil
}

// DeleteI2PKeyStore deletes the I2P Keystore.
func DeleteI2PKeyStore() error {
	log.WithField("path", I2P_KEYSTORE_PATH).Debug("Attempting to delete I2P keystore")
	err := os.RemoveAll(I2P_KEYSTORE_PATH)
	if err != nil {
		log.WithError(err).WithField("path", I2P_KEYSTORE_PATH).Error("Failed to delete I2P keystore")
		return err
	}
	log.WithField("path", I2P_KEYSTORE_PATH).Debug("Successfully deleted I2P keystore")
	return nil
	//return os.RemoveAll(I2P_KEYSTORE_PATH)
}

// TorKeystorePath returns the path to the Onion Keystore. If the
// path is not set, it returns the default path. If the path does
// not exist, it creates it.
func TorKeystorePath() (string, error) {
	log.WithField("path", ONION_KEYSTORE_PATH).Debug("Checking Tor keystore path")
	if _, err := os.Stat(ONION_KEYSTORE_PATH); err != nil {
		log.WithField("path", ONION_KEYSTORE_PATH).Debug("Tor keystore directory does not exist, creating")
		err := os.MkdirAll(ONION_KEYSTORE_PATH, 0755)
		if err != nil {
			log.WithError(err).WithField("path", ONION_KEYSTORE_PATH).Error("Failed to create Tor keystore directory")
			return "", err
		}
	}
	log.WithField("path", ONION_KEYSTORE_PATH).Debug("Tor keystore path verified")
	return ONION_KEYSTORE_PATH, nil
}

// DeleteTorKeyStore deletes the Onion Keystore.
func DeleteTorKeyStore() error {
	log.WithField("path", ONION_KEYSTORE_PATH).Debug("Attempting to delete Tor keystore")
	err := os.RemoveAll(ONION_KEYSTORE_PATH)
	if err != nil {
		log.WithError(err).WithField("path", ONION_KEYSTORE_PATH).Error("Failed to delete Tor keystore")
		return err
	}
	log.WithField("path", ONION_KEYSTORE_PATH).Debug("Successfully deleted Tor keystore")
	return nil
	//return os.RemoveAll(ONION_KEYSTORE_PATH)
}

// TLSKeystorePath returns the path to the TLS Keystore. If the
// path is not set, it returns the default path. If the path does
// not exist, it creates it.
func TLSKeystorePath() (string, error) {
	log.WithField("path", TLS_KEYSTORE_PATH).Debug("Checking TLS keystore path")
	if _, err := os.Stat(TLS_KEYSTORE_PATH); err != nil {
		log.WithField("path", TLS_KEYSTORE_PATH).Debug("TLS keystore directory does not exist, creating")
		err := os.MkdirAll(TLS_KEYSTORE_PATH, 0755)
		if err != nil {
			log.WithError(err).WithField("path", TLS_KEYSTORE_PATH).Error("Failed to create TLS keystore directory")
			return "", err
		}
	}
	log.WithField("path", TLS_KEYSTORE_PATH).Debug("TLS keystore path verified")
	return TLS_KEYSTORE_PATH, nil
}

// DeleteTLSKeyStore deletes the TLS Keystore.
func DeleteTLSKeyStore() error {
	log.WithField("path", TLS_KEYSTORE_PATH).Debug("Attempting to delete TLS keystore")
	err := os.RemoveAll(TLS_KEYSTORE_PATH)
	if err != nil {
		log.WithError(err).WithField("path", TLS_KEYSTORE_PATH).Error("Failed to delete TLS keystore")
		return err
	}
	log.WithField("path", TLS_KEYSTORE_PATH).Debug("Successfully deleted TLS keystore")
	return nil
	//return os.RemoveAll(TLS_KEYSTORE_PATH)
}

// Dial returns a connection for the given network and address.
// network is ignored. If the address ends in i2p, it returns an I2P connection.
// if the address ends in anything else, it returns a Tor connection.
func Dial(network, addr string) (net.Conn, error) {
	log.WithFields(logrus.Fields{
		"network": network,
		"address": addr,
	}).Debug("Attempting to dial")

	url, err := url.Parse(addr)
	if err != nil {
		log.WithError(err).WithField("address", addr).Error("Failed to parse address")
		return nil, err
	}
	hostname := url.Hostname()
	if strings.HasSuffix(hostname, ".i2p") {
		log.WithField("hostname", hostname).Debug("Using I2P connection for .i2p address")
		return DialGarlic(network, addr)
	}
	log.WithField("hostname", hostname).Debug("Using Tor connection for non-i2p address")
	return DialOnion(network, addr)
}

// Listen returns a listener for the given network and address.
// if network is i2p or garlic, it returns an I2P listener.
// if network is tor or onion, it returns an Onion listener.
// if keys ends with ".i2p", it returns an I2P listener.
func Listen(network, keys string) (net.Listener, error) {
	log.WithFields(logrus.Fields{
		"network": network,
		"keys":    keys,
	}).Debug("Attempting to create listener")

	if network == "i2p" || network == "garlic" {
		log.Debug("Creating I2P listener based on network type")
		return ListenGarlic(network, keys)
	}
	if network == "tor" || network == "onion" {
		log.Debug("Creating Tor listener based on network type")
		return ListenOnion(network, keys)
	}

	url, err := url.Parse(keys)
	if err != nil {
		log.WithError(err).WithField("keys", keys).Error("Failed to parse keys URL")
		return nil, err
	}

	hostname := url.Hostname()
	if strings.HasSuffix(hostname, ".i2p") {
		log.WithField("hostname", hostname).Debug("Creating I2P listener based on .i2p hostname")
		return ListenGarlic(network, keys)
	}
	log.WithField("hostname", hostname).Debug("Creating Tor listener for non-i2p hostname")
	return ListenOnion(network, keys)
}
