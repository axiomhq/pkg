package cmd

import "go.uber.org/zap"

// DefaultLoggerOptions are the default logger options to use
func DefaultLoggerOptions() []zap.Option {
	return []zap.Option{
		zap.AddStacktrace(zap.DPanicLevel),
	}
}
