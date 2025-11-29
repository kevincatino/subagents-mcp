package logging

import (
	"go.uber.org/zap"
)

// New returns a production zap logger configured for JSON output.
func New() (*zap.Logger, error) {
	return zap.NewProduction()
}
