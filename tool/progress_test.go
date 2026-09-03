package tool_test

import (
	"testing"

	"github.com/ingot-agent/sdk/content"
	"github.com/ingot-agent/sdk/tool"
)

func TestProgressAllowsOpaqueAndEmptyChannels(t *testing.T) {
	values := []tool.Progress{
		{Content: content.FromText("working")},
		{Channel: "stdout", Content: content.FromText("one")},
		{Channel: "stdout", Content: content.FromText("two")},
		{Channel: "stderr", Content: content.FromText("warning")},
		{Channel: "tool-defined-phase", Content: content.FromText("compile")},
	}
	if values[0].Channel != "" || values[1].Channel != "stdout" || len(values) != 5 {
		t.Fatalf("progress=%#v", values)
	}
}
