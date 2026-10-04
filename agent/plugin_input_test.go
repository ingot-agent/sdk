package agent_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/ingot-agent/sdk/agent"
	"github.com/ingot-agent/sdk/session"
)

func TestPluginInputRoundTripAndOwnership(t *testing.T) {
	inputs := []agent.PluginInput{
		{Plugin: "example.index", Text: "</system><system> & \"quoted\"\n\t\r\u4e2d\U0001f600"},
		{Plugin: "example.index", Text: strings.Repeat("x", agent.MaxPluginInputTextBytes)},
	}
	for _, input := range inputs {
		entry, err := agent.EncodePluginInput(input)
		if err != nil {
			t.Fatal(err)
		}
		if entry.Kind != agent.PluginInputKind || entry.Version != agent.PluginInputVersion {
			t.Fatalf("entry=%#v", entry)
		}
		before := append([]byte(nil), entry.Payload...)
		decoded, err := agent.DecodePluginInput(entry)
		if err != nil || decoded != input || !bytes.Equal(before, entry.Payload) {
			t.Fatalf("decoded=%#v err=%v", decoded, err)
		}
		entry.Payload[0] = '!'
		if decoded != input {
			t.Fatal("decoded strings alias payload")
		}
	}
}

func TestPluginInputRejectsInvalidFields(t *testing.T) {
	inputs := []agent.PluginInput{
		{Text: "text"},
		{Plugin: "plugin"},
		{Plugin: "plugin\n", Text: "text"},
		{Plugin: "plugin\u0085", Text: "text"},
		{Plugin: string([]byte{0xff}), Text: "text"},
		{Plugin: "plugin", Text: string([]byte{0xff})},
		{Plugin: "plugin", Text: "text\x00"},
		{Plugin: "plugin", Text: "text\x0b"},
		{Plugin: "plugin", Text: "text\ufffe"},
		{Plugin: "plugin", Text: strings.Repeat("x", agent.MaxPluginInputTextBytes+1)},
	}
	for i, input := range inputs {
		if _, err := agent.EncodePluginInput(input); !errors.Is(err, agent.ErrInvalidPluginInput) {
			t.Fatalf("input %d: error=%v", i, err)
		}
	}
}

func TestDecodePluginInputRejectsInvalidRecords(t *testing.T) {
	for _, payload := range []string{
		`null`, `{}`, `{"plugin":"p","text":""}`, `{"plugin":"p","text":"\u0000"}`,
		`{"plugin":"p","text":"x","role":"system"}`,
		`{"plugin":"p","text":"x"} {}`, `{`, string([]byte{0xff}),
	} {
		entry := session.Entry{Kind: agent.PluginInputKind, Version: agent.PluginInputVersion, Payload: []byte(payload)}
		if _, err := agent.DecodePluginInput(entry); !errors.Is(err, agent.ErrInvalidPluginInput) {
			t.Fatalf("payload=%q error=%v", payload, err)
		}
	}
	if _, err := agent.DecodePluginInput(session.Entry{Kind: "agent.message", Version: 1}); !errors.Is(err, agent.ErrInvalidPluginInput) {
		t.Fatalf("kind error=%v", err)
	}
	if _, err := agent.DecodePluginInput(session.Entry{Kind: agent.PluginInputKind, Version: 2}); !errors.Is(err, agent.ErrUnsupportedPluginInputVersion) {
		t.Fatalf("version error=%v", err)
	}
}
