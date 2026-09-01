package tool

import "github.com/ingot-agent/sdk/content"

// Progress is one transient incremental fact produced while a Tool invocation
// is still running. Channel is an opaque, tool-defined logical stream name; an
// empty channel means the Tool does not distinguish progress streams.
//
// Progress is not a partial Result, and its Content is not required to be
// reproduced in the invocation's final Result. Aggregate inputs are immutable
// by contract; consumers that retain Content must make a deep copy.
type Progress struct {
	Channel string
	Content content.Content
}
