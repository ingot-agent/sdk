package content_test

import (
	"errors"
	"reflect"
	"testing"
	"unicode/utf8"

	"github.com/ingot-agent/sdk/asset"
	"github.com/ingot-agent/sdk/content"
)

func TestConstructCloneAndNormalize(t *testing.T) {
	t.Parallel()
	data := []byte{1, 2, 3}
	image := content.Inline(content.KindImage, "image/png", "diagram.png", data)
	data[0] = 9
	if image.Media.Source.Data[0] != 1 {
		t.Fatal("Inline retained caller data")
	}

	attachments := []content.Attachment{{Kind: image.Kind, Media: image.Media}}
	got := content.FromInput("describe", attachments)
	want := content.Content{content.Text("describe"), image}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FromInput = %#v, want %#v", got, want)
	}
	attachments[0].Media.Source.Data[1] = 8
	if got[1].Media.Source.Data[1] != 2 {
		t.Fatal("FromInput retained attachment data")
	}

	cloned := content.Clone(got)
	cloned[1].Media.Source.Data[2] = 7
	if got[1].Media.Source.Data[2] != 3 {
		t.Fatal("Clone retained inline data")
	}
	if got := content.FromInput("", attachments); len(got) != 1 || got[0].Kind != content.KindImage {
		t.Fatalf("attachment-only input = %#v", got)
	}
	if got := content.FromInput("", nil); len(got) != 0 {
		t.Fatalf("empty input = %#v", got)
	}
}

func TestTextOnly(t *testing.T) {
	t.Parallel()
	if got, ok := content.TextOnly(nil); !ok || got != "" {
		t.Fatalf("empty TextOnly = %q, %v", got, ok)
	}
	if got, ok := content.TextOnly(content.Content{content.Text("a"), content.Text(""), content.Text("b")}); !ok || got != "ab" {
		t.Fatalf("TextOnly = %q, %v", got, ok)
	}
	if got, ok := content.TextOnly(content.Content{content.Text("a"), content.URI(content.KindImage, "", "", "not validated")}); ok || got != "" {
		t.Fatalf("media TextOnly = %q, %v", got, ok)
	}
}

func TestValidateAcceptedEmptyAndOpaqueValues(t *testing.T) {
	t.Parallel()
	invalidURI := string([]byte{0xff})
	valid := []content.Content{
		nil,
		{},
		{content.Text("")},
		{content.Inline(content.KindImage, "", "", nil)},
		{content.Inline(content.KindAudio, "NOT A MIME TYPE", "", []byte{})},
		{content.URI(content.KindVideo, "video/example", "", invalidURI)},
		{content.AssetPart(content.KindFile, "application/pdf", "report.pdf", asset.Reference{ID: "opaque"})},
	}
	for i, value := range valid {
		if err := content.Validate(value); err != nil {
			t.Fatalf("valid[%d]: %v", i, err)
		}
	}
}

func TestValidateRejectsInvalidUnions(t *testing.T) {
	t.Parallel()
	invalidUTF8 := string([]byte{utf8.RuneSelf})
	tests := []content.Content{
		{{}},
		{{Kind: 99}},
		{{Kind: content.KindText, Text: invalidUTF8}},
		{{Kind: content.KindText, Media: content.Media{Name: "media"}}},
		{{Kind: content.KindImage, Text: "text", Media: content.Media{Source: content.Source{Kind: content.SourceInline}}}},
		{{Kind: content.KindImage, Media: content.Media{Name: invalidUTF8, Source: content.Source{Kind: content.SourceInline}}}},
		{{Kind: content.KindImage, Media: content.Media{Source: content.Source{Kind: 99}}}},
		{{Kind: content.KindImage, Media: content.Media{Source: content.Source{Kind: content.SourceInline, URI: "uri"}}}},
		{{Kind: content.KindImage, Media: content.Media{Source: content.Source{Kind: content.SourceURI, Data: []byte{1}}}}},
		{{Kind: content.KindImage, Media: content.Media{Source: content.Source{Kind: content.SourceAsset}}}},
		{{Kind: content.KindImage, Media: content.Media{Source: content.Source{Kind: content.SourceAsset, Asset: asset.Reference{ID: "id"}, URI: "uri"}}}},
	}
	for i, value := range tests {
		err := content.Validate(value)
		if !errors.Is(err, content.ErrInvalidContent) {
			t.Fatalf("invalid[%d] error = %v", i, err)
		}
	}
	if err := content.ValidateAttachments([]content.Attachment{{Kind: content.KindText}}); !errors.Is(err, content.ErrInvalidContent) {
		t.Fatalf("text attachment error = %v", err)
	}
}

func TestAttachmentHelpersOwnInlineData(t *testing.T) {
	t.Parallel()
	attachments := []content.Attachment{{
		Kind:  content.KindImage,
		Media: content.Media{Source: content.Source{Kind: content.SourceInline, Data: []byte{1}}},
	}}
	if err := content.ValidateAttachments(attachments); err != nil {
		t.Fatal(err)
	}
	part := content.AttachmentPart(attachments[0])
	clone := content.CloneAttachments(attachments)
	part.Media.Source.Data[0] = 2
	clone[0].Media.Source.Data[0] = 3
	if attachments[0].Media.Source.Data[0] != 1 {
		t.Fatal("attachment helper retained inline data")
	}
}

func TestUnsupportedError(t *testing.T) {
	t.Parallel()
	err := &content.UnsupportedError{MessageIndex: 2, PartIndex: 3, Kind: content.KindImage, MIMEType: "image/png", Reason: "source"}
	if !errors.Is(err, content.ErrUnsupportedContent) {
		t.Fatalf("errors.Is(%v) = false", err)
	}
	var target *content.UnsupportedError
	if !errors.As(err, &target) || target != err {
		t.Fatalf("errors.As(%v) = %#v", err, target)
	}
}
