// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-multi-json/multi-json authors

package multijson

// ParseError models MultiJson::ParseError (aliased in the gem as
// MultiJson::LoadError and MultiJson::DecodeError): the error raised when an
// adapter fails to parse a JSON string. It carries the offending input in Data
// and the underlying engine error in Cause, mirroring the gem's #data and
// #cause readers.
type ParseError struct {
	// Data is the input string that failed to parse (the gem's #data).
	Data string
	// Cause is the underlying engine error (the gem's #cause).
	Cause error
}

// Error implements the error interface, reporting the underlying cause's
// message, as MultiJson::ParseError forwards the wrapped exception's message.
func (e *ParseError) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "multijson: parse error"
}

// Unwrap exposes the underlying cause so errors.Is/errors.As can reach it.
func (e *ParseError) Unwrap() error { return e.Cause }

// LoadError is the alias the gem defines as MultiJson::LoadError = ParseError.
type LoadError = ParseError

// AdapterError models MultiJson::AdapterError, the ArgumentError the gem raises
// when an adapter cannot be recognised or loaded. Name is the rejected adapter
// specification and Cause, when present, is the underlying failure.
type AdapterError struct {
	// Name is the adapter specification that could not be resolved.
	Name string
	// Cause is the underlying failure, if any (the gem wraps a LoadError/NameError).
	Cause error
}

// Error implements the error interface with a message shaped like the gem's
// AdapterError.build output.
func (e *AdapterError) Error() string {
	msg := "multijson: did not recognize your adapter specification (" + e.Name + ")"
	if e.Cause != nil {
		return msg + ": " + e.Cause.Error()
	}
	return msg
}

// Unwrap exposes the underlying cause, if any.
func (e *AdapterError) Unwrap() error { return e.Cause }
