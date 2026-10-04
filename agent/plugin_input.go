package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/ingot-agent/sdk/session"
)

const (
	// PluginInputKind identifies a durable plugin-provided context record.
	PluginInputKind = "agent.plugin_input"
	// PluginInputVersion identifies the PluginInput JSON payload schema.
	PluginInputVersion = 1
	// MaxPluginInputTextBytes bounds the unwrapped UTF-8 text of one record.
	MaxPluginInputTextBytes = 64 * 1024
)

var (
	// ErrInvalidPluginInput indicates invalid fields, kind, or JSON payload.
	ErrInvalidPluginInput = errors.New("invalid plugin input")
	// ErrUnsupportedPluginInputVersion indicates an unknown payload schema.
	ErrUnsupportedPluginInputVersion = errors.New("unsupported plugin input version")
)

// PluginInput is text supplied by a plugin for one session's context. Plugin is
// the manifest's short name, supplied by the caller without authentication.
// Text is nonempty UTF-8, at most MaxPluginInputTextBytes, and representable as
// XML 1.0 text. Plugin must be nonempty UTF-8 without control characters.
//
// Agents supporting this record project it to a user-role message enclosed in
// <system source="plugin" plugin="...">...</system>. The protocol defines no
// business meaning, instruction priority, or model response requirements.
type PluginInput struct {
	Plugin string `json:"plugin"`
	Text   string `json:"text"`
}

// EncodePluginInput returns a caller-owned opaque Entry for Store.Append.
// Successful Append means the record is committed, not that a model has seen
// it. Inclusion in a particular Turn or Round is not guaranteed. Appending does
// not start a Turn or change an existing request snapshot. Store's commit and
// error semantics apply; this protocol provides no retry or deduplication layer.
func EncodePluginInput(input PluginInput) (session.Entry, error) {
	if err := validatePluginInput(input); err != nil {
		return session.Entry{}, err
	}
	payload, err := json.Marshal(input)
	if err != nil {
		return session.Entry{}, fmt.Errorf("encode plugin input: %w", err)
	}
	return session.Entry{Kind: PluginInputKind, Version: PluginInputVersion, Payload: payload}, nil
}

// DecodePluginInput validates and decodes one Entry without retaining or
// modifying its Payload. Other kinds return ErrInvalidPluginInput; unsupported
// versions return ErrUnsupportedPluginInputVersion. Invalid JSON, unknown
// fields, multiple JSON values, or invalid fields return ErrInvalidPluginInput.
func DecodePluginInput(entry session.Entry) (PluginInput, error) {
	if entry.Kind != PluginInputKind {
		return PluginInput{}, fmt.Errorf("entry kind %q: %w", entry.Kind, ErrInvalidPluginInput)
	}
	if entry.Version != PluginInputVersion {
		return PluginInput{}, fmt.Errorf("entry version %d: %w", entry.Version, ErrUnsupportedPluginInputVersion)
	}
	if !utf8.Valid(entry.Payload) {
		return PluginInput{}, fmt.Errorf("payload is not valid UTF-8: %w", ErrInvalidPluginInput)
	}
	decoder := json.NewDecoder(bytes.NewReader(entry.Payload))
	decoder.DisallowUnknownFields()
	var input PluginInput
	if err := decoder.Decode(&input); err != nil {
		return PluginInput{}, fmt.Errorf("decode plugin input: %w: %w", ErrInvalidPluginInput, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return PluginInput{}, fmt.Errorf("payload must contain one JSON value: %w", ErrInvalidPluginInput)
	}
	if err := validatePluginInput(input); err != nil {
		return PluginInput{}, err
	}
	return input, nil
}

func validatePluginInput(input PluginInput) error {
	if input.Plugin == "" || !utf8.ValidString(input.Plugin) || !validPluginInputXML(input.Plugin) {
		return fmt.Errorf("plugin name: %w", ErrInvalidPluginInput)
	}
	for _, r := range input.Plugin {
		if unicode.IsControl(r) {
			return fmt.Errorf("plugin name contains control characters: %w", ErrInvalidPluginInput)
		}
	}
	if input.Text == "" || len(input.Text) > MaxPluginInputTextBytes || !utf8.ValidString(input.Text) || !validPluginInputXML(input.Text) {
		return fmt.Errorf("text must be nonempty XML-compatible UTF-8 and at most %d bytes: %w", MaxPluginInputTextBytes, ErrInvalidPluginInput)
	}
	return nil
}

func validPluginInputXML(text string) bool {
	for _, r := range text {
		if r == '\t' || r == '\n' || r == '\r' ||
			(r >= 0x20 && r <= 0xD7FF) ||
			(r >= 0xE000 && r <= 0xFFFD) ||
			(r >= 0x10000 && r <= utf8.MaxRune) {
			continue
		}
		return false
	}
	return true
}
