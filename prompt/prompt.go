// Package prompt defines prompt contributors and renderers.
package prompt

import (
	"context"

	"github.com/ingot-agent/sdk/content"
	"github.com/ingot-agent/sdk/model"
	"github.com/ingot-agent/sdk/session"
)

// Request is the input to contributors and renderers.
type Request struct {
	SessionID session.ID
	Input     content.Content
	History   []model.Message
}

// Block is one named prompt contribution.
type Block struct {
	Name    string
	Content content.Content
}

// Contributor produces prompt blocks. Collections of contributors are invoked
// in component MANY aggregation order by the renderer implementation.
type Contributor interface {
	Contribute(context.Context, Request) ([]Block, error)
}

// Renderer creates the messages sent to a model.
type Renderer interface {
	Render(context.Context, Request) ([]model.Message, error)
}
