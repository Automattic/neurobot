package matrix

import (
	"fmt"

	"github.com/apex/log"
	"maunium.net/go/mautrix/crypto"
)

type apexLogger struct{}

func (f apexLogger) Error(message string, args ...interface{}) {
	log.Error(fmt.Sprintf(message, args...))
}

func (f apexLogger) Warn(message string, args ...interface{}) {
	log.Warn(fmt.Sprintf(message, args...))
}

func (f apexLogger) Debug(message string, args ...interface{}) {
	log.Debug(fmt.Sprintf(message, args...))
}

func (f apexLogger) Trace(message string, args ...interface{}) {
	log.Trace(fmt.Sprintf(message, args...))
}

// NewApexLogger returns an instance of a wrapper around apex logger that satisfies the interface needed by olm machine logger
func NewApexLogger() crypto.Logger {
	return &apexLogger{}
}
