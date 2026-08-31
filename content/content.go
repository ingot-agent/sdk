// Package content defines ordered multimodal content shared by general agent
// capabilities.
package content

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/ingot-agent/sdk/asset"
)

// Kind identifies the semantic modality of a Part.
type Kind uint8

const (
	KindText Kind = iota + 1
	KindImage
	KindAudio
	KindVideo
	KindFile
)

// Content is an ordered sequence of content parts. Part order and boundaries
// are significant and must be preserved by consumers.
type Content []Part

// Part is one text or media item in a Content sequence. Kind selects the
// authoritative representation; all inactive representation fields must be
// zero.
type Part struct {
	Kind  Kind
	Text  string
	Media Media
}

// Media describes one non-text content item. MIMEType and Name are descriptive
// caller-supplied values; the SDK does not inspect the represented bytes.
type Media struct {
	MIMEType string
	Name     string
	Source   Source
}

// Attachment is one non-text input attached to an agent turn. Attachments are
// ordered by slice position; that order does not establish references to
// substrings within the turn input.
type Attachment struct {
	Kind  Kind
	Media Media
}

// SourceKind identifies how media bytes are made available.
type SourceKind uint8

const (
	SourceInline SourceKind = iota + 1
	SourceURI
	SourceAsset
)

// Source carries exactly one media representation selected by Kind. A URI is
// only a location or identifier and does not grant a consumer permission to
// read it.
type Source struct {
	Kind  SourceKind
	Data  []byte
	URI   string
	Asset asset.Reference
}

var (
	// ErrInvalidContent indicates that content violates the structural tagged
	// union or text encoding contract.
	ErrInvalidContent = errors.New("invalid content")
	// ErrUnsupportedContent indicates that a valid content part cannot be
	// handled by a selected implementation.
	ErrUnsupportedContent = errors.New("unsupported content")
)

// UnsupportedError identifies a content part rejected by an implementation.
// MessageIndex and PartIndex identify its position when known.
type UnsupportedError struct {
	MessageIndex int
	PartIndex    int
	Kind         Kind
	MIMEType     string
	Reason       string
}

// Error describes the unsupported part without exposing inline data or URI
// values.
func (e *UnsupportedError) Error() string {
	if e == nil {
		return ErrUnsupportedContent.Error()
	}
	if e.Reason == "" {
		return fmt.Sprintf("%s: message %d part %d kind %d MIME type %q", ErrUnsupportedContent, e.MessageIndex, e.PartIndex, e.Kind, e.MIMEType)
	}
	return fmt.Sprintf("%s: message %d part %d kind %d MIME type %q: %s", ErrUnsupportedContent, e.MessageIndex, e.PartIndex, e.Kind, e.MIMEType, e.Reason)
}

// Unwrap makes UnsupportedError match ErrUnsupportedContent through errors.Is.
func (e *UnsupportedError) Unwrap() error { return ErrUnsupportedContent }

// Text constructs a text part.
func Text(value string) Part { return Part{Kind: KindText, Text: value} }

// Inline constructs a media part and copies data before returning.
func Inline(kind Kind, mimeType, name string, data []byte) Part {
	return Part{
		Kind: kind,
		Media: Media{
			MIMEType: mimeType,
			Name:     name,
			Source: Source{
				Kind: SourceInline,
				Data: cloneBytes(data),
			},
		},
	}
}

// URI constructs a media part that refers to a URI. The URI is preserved
// exactly and is not parsed, validated, normalized, or read by this package.
func URI(kind Kind, mimeType, name, uri string) Part {
	return Part{
		Kind: kind,
		Media: Media{
			MIMEType: mimeType,
			Name:     name,
			Source:   Source{Kind: SourceURI, URI: uri},
		},
	}
}

// AssetPart constructs a media part backed by an opaque asset reference. It
// does not resolve or read the asset.
func AssetPart(kind Kind, mimeType, name string, reference asset.Reference) Part {
	return Part{
		Kind: kind,
		Media: Media{
			MIMEType: mimeType,
			Name:     name,
			Source:   Source{Kind: SourceAsset, Asset: reference},
		},
	}
}

// FromText constructs single-part text content.
func FromText(value string) Content { return Content{Text(value)} }

// FromInput normalizes an agent input as an optional leading text part followed
// by attachments in slice order. Inline data is copied before returning.
func FromInput(input string, attachments []Attachment) Content {
	result := make(Content, 0, len(attachments)+1)
	if input != "" {
		result = append(result, Text(input))
	}
	for _, attachment := range attachments {
		result = append(result, AttachmentPart(attachment))
	}
	return result
}

// Clone returns a caller-owned deep copy, including every inline byte slice.
func Clone(value Content) Content {
	if value == nil {
		return nil
	}
	result := make(Content, len(value))
	for i, part := range value {
		result[i] = clonePart(part)
	}
	return result
}

// Validate checks the structural tagged-union and UTF-8 contracts. It does not
// validate MIME type strings, URI strings, media bytes, or implementation
// policy such as supported modalities and size limits.
func Validate(value Content) error {
	for i, part := range value {
		if err := validatePart(part); err != nil {
			return fmt.Errorf("%w: part %d: %v", ErrInvalidContent, i, err)
		}
	}
	return nil
}

// TextOnly joins all text parts directly in order. It succeeds only when every
// part is text; it never skips media parts or inserts separators.
func TextOnly(value Content) (string, bool) {
	var result string
	for _, part := range value {
		if part.Kind != KindText {
			return "", false
		}
		result += part.Text
	}
	return result, true
}

// CloneAttachments returns a caller-owned deep copy, including every inline
// byte slice.
func CloneAttachments(value []Attachment) []Attachment {
	if value == nil {
		return nil
	}
	result := make([]Attachment, len(value))
	for i, attachment := range value {
		result[i] = attachment
		result[i].Media.Source.Data = cloneBytes(attachment.Media.Source.Data)
	}
	return result
}

// ValidateAttachments validates attachments in slice order without reordering
// them.
func ValidateAttachments(value []Attachment) error {
	for i, attachment := range value {
		if attachment.Kind == KindText {
			return fmt.Errorf("%w: attachment %d: text kind is not allowed", ErrInvalidContent, i)
		}
		if err := validatePart(AttachmentPart(attachment)); err != nil {
			return fmt.Errorf("%w: attachment %d: %v", ErrInvalidContent, i, err)
		}
	}
	return nil
}

// AttachmentPart performs a lossless structural conversion and copies inline
// data. It does not resolve URI or asset sources.
func AttachmentPart(value Attachment) Part {
	return clonePart(Part{Kind: value.Kind, Media: value.Media})
}

func validatePart(part Part) error {
	switch part.Kind {
	case KindText:
		if !utf8.ValidString(part.Text) {
			return errors.New("text is not valid UTF-8")
		}
		if part.Media.MIMEType != "" || part.Media.Name != "" || part.Media.Source.Kind != 0 || len(part.Media.Source.Data) != 0 || part.Media.Source.URI != "" || part.Media.Source.Asset.ID != "" {
			return errors.New("text part carries media fields")
		}
		return nil
	case KindImage, KindAudio, KindVideo, KindFile:
		if part.Text != "" {
			return errors.New("media part carries text")
		}
		if !utf8.ValidString(part.Media.Name) {
			return errors.New("media name is not valid UTF-8")
		}
		return validateSource(part.Media.Source)
	default:
		return fmt.Errorf("unknown kind %d", part.Kind)
	}
}

func validateSource(source Source) error {
	switch source.Kind {
	case SourceInline:
		if source.URI != "" || source.Asset.ID != "" {
			return errors.New("inline source carries another representation")
		}
	case SourceURI:
		if len(source.Data) != 0 || source.Asset.ID != "" {
			return errors.New("URI source carries another representation")
		}
	case SourceAsset:
		if source.Asset.ID == "" {
			return errors.New("asset source has an empty reference")
		}
		if len(source.Data) != 0 || source.URI != "" {
			return errors.New("asset source carries another representation")
		}
	default:
		return fmt.Errorf("unknown source kind %d", source.Kind)
	}
	return nil
}

func clonePart(value Part) Part {
	result := value
	result.Media.Source.Data = cloneBytes(value.Media.Source.Data)
	return result
}

func cloneBytes(value []byte) []byte {
	if value == nil {
		return nil
	}
	return append([]byte(nil), value...)
}
