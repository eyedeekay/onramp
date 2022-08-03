package onramp

import (
	"os"
	"path/filepath"
)

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

// I2P_KEYSTORE_PATH is the place where I2P Keys will be saved.
// it defaults to the directory "i2pkeys" current working directory
// reference it by calling I2PKeystorePath() to check for errors
var I2P_KEYSTORE_PATH = i2pdefault

// ONION_KEYSTORE_PATH is the place where Onion Keys will be saved.
// it defaults to the directory "onionkeys" current working directory
// reference it by calling OnionKeystorePath() to check for errors
var ONION_KEYSTORE_PATH = tordefault

func I2PKeystorePath() (string, error) {
	if _, err := os.Stat(I2P_KEYSTORE_PATH); err != nil {
		err := os.MkdirAll(I2P_KEYSTORE_PATH, 0755)
		if err != nil {
			return "", err
		}
	}
	return I2P_KEYSTORE_PATH, nil
}

func TorKeystorePath() (string, error) {
	if _, err := os.Stat(ONION_KEYSTORE_PATH); err != nil {
		err := os.MkdirAll(ONION_KEYSTORE_PATH, 0755)
		if err != nil {
			return "", err
		}
	}
	return ONION_KEYSTORE_PATH, nil
}
