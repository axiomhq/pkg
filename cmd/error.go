package cmd

import (
	"fmt"

	"go.uber.org/zap"
)

// mainFuncError is an error returned by the `MainFunc` that enhances logging
// output. In should be created using the `Error()` function.
type mainFuncError struct {
	msg    string
	err    error
	fields []zap.Field
}

// Error implements `error`.
func (mfe *mainFuncError) Error() string {
	return fmt.Errorf("%s: %w", mfe.msg, mfe.err).Error()
}

// Fields returns all `zap.Field` including the error.
func (mfe *mainFuncError) Fields() []zap.Field {
	return append(mfe.fields, zap.Error(mfe.err))
}

// Error is a convenience function that improves error log output when returning
// from the `MainFunc`.
func Error(msg string, err error, fields ...zap.Field) error {
	return &mainFuncError{msg, err, fields}
}
