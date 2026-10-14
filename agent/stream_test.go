package agent_test

import (
	"github.com/ingot-agent/sdk/agent"
	"github.com/ingot-agent/sdk/content"
	"testing"
)

func TestInputCloneOwnsMediaAndFiles(t *testing.T) {
	input := agent.Input{Text: "caption", Attachments: agent.MediaAttachments([]content.Attachment{{Kind: content.KindImage, Media: content.Inline(content.KindImage, "image/png", "image", []byte{1, 2}).Media}})}
	input.Attachments = append(input.Attachments, agent.InputAttachment{Path: "notes.md", Name: "notes.md", Size: 12})
	if err := agent.ValidateInput(input); err != nil {
		t.Fatal(err)
	}
	turn := agent.Turn{SessionID: "s", Inputs: []agent.Input{input}}
	copy := agent.CloneTurn(turn)
	copy.Inputs[0].Text = "changed"
	copy.Inputs[0].Attachments[0].Media.Media.Source.Data[0] = 9
	copy.Inputs[0].Attachments[1].Path = "changed"
	if turn.Inputs[0].Text != "caption" || input.Attachments[0].Media.Media.Source.Data[0] != 1 || input.Attachments[1].Path != "notes.md" {
		t.Fatal("clone aliases input")
	}
	for _, bad := range []agent.Input{{Text: string([]byte{0xff})}, {Attachments: []agent.InputAttachment{{}}}, {Attachments: []agent.InputAttachment{{Path: "a", Size: -1}}}} {
		if agent.ValidateInput(bad) == nil {
			t.Fatal("accepted invalid input")
		}
	}
}
