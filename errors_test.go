// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-multi-json/multi-json authors

package multijson

import (
	"errors"
	"strings"
	"testing"
)

func TestParseErrorMessageWithoutCause(t *testing.T) {
	pe := &ParseError{Data: "oops"}
	if pe.Error() != "multijson: parse error" {
		t.Errorf("Error() = %q", pe.Error())
	}
	if pe.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", pe.Unwrap())
	}
}

func TestParseErrorMessageWithCause(t *testing.T) {
	cause := errors.New("bad token")
	pe := &ParseError{Data: "x", Cause: cause}
	if pe.Error() != "bad token" {
		t.Errorf("Error() = %q", pe.Error())
	}
	if !errors.Is(pe, cause) {
		t.Error("errors.Is could not reach the cause")
	}
}

func TestLoadErrorAlias(t *testing.T) {
	// LoadError is a type alias for ParseError.
	var le *LoadError = &ParseError{Data: "d"}
	if le.Data != "d" {
		t.Errorf("alias mismatch: %q", le.Data)
	}
}

func TestAdapterErrorMessageWithoutCause(t *testing.T) {
	ae := &AdapterError{Name: "nope"}
	if !strings.Contains(ae.Error(), "nope") {
		t.Errorf("Error() = %q, want to mention name", ae.Error())
	}
	if ae.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", ae.Unwrap())
	}
}

func TestAdapterErrorMessageWithCause(t *testing.T) {
	cause := errors.New("underlying")
	ae := &AdapterError{Name: "nope", Cause: cause}
	msg := ae.Error()
	if !strings.Contains(msg, "nope") || !strings.Contains(msg, "underlying") {
		t.Errorf("Error() = %q", msg)
	}
	if !errors.Is(ae, cause) {
		t.Error("errors.Is could not reach the cause")
	}
}
