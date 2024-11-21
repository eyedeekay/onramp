package onramp

import (
	"github.com/go-i2p/logger"
)

var (
	log  *logger.Logger
)

func InitializeOnrampLogger() {
	logger.InitializeGoI2PLogger()
	log = logger.GetGoI2PLogger()
}

// GetI2PKeysLogger returns the initialized logger
func GetOnrampLogger() *logger.Logger {
	return logger.GetGoI2PLogger()
}

func init() {
	GetOnrampLogger()
}
