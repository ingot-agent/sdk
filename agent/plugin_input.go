package agent

import (
	"context"

	"github.com/ingot-agent/sdk/model"
	"github.com/ingot-agent/sdk/session"
)

// PluginInput is text supplied by a plugin for one session's context. Plugin
// identifies the caller without authentication. Providers own validation,
// persistence formats, and message envelopes. No instruction priority or model
// response requirement is defined by this contract.
type PluginInput struct {
	Plugin string `json:"plugin"`
	Text   string `json:"text"`
}

// PluginInputWriter appends input to an existing session. A successful call
// means the input is committed, not that a model has seen it. Inclusion in a
// particular Turn or Round is not guaranteed. Appending does not start a Turn
// or change an existing request snapshot. Store's commit and error semantics
// apply; callers must not retry merely because an error was returned.
type PluginInputWriter interface {
	Append(context.Context, session.ID, PluginInput) error
}

// PluginInputProjector recognizes provider-owned records and projects them to
// caller-owned user-role, text-only messages without modifying the Entry.
// Unknown records return recognized=false and a nil error. Recognized invalid
// records return recognized=true and an error; callers must propagate it.
type PluginInputProjector interface {
	Project(session.Entry) (message model.Message, recognized bool, err error)
}
