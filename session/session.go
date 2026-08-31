// Package session defines append-oriented session persistence.
package session

import (
	"context"
	"errors"
	"time"
)

// ID identifies one session.
type ID string

// Metadata is supplied when creating a session.
type Metadata struct {
	Title     string
	CreatedAt time.Time
}

// Entry is a durable, versioned persistence envelope. Version identifies the
// payload schema for Kind.
type Entry struct {
	Kind    string
	Version int
	Payload []byte
}

// Query controls session listing pagination.
type Query struct {
	Limit  int
	Offset int
}

// Summary describes a persisted session.
type Summary struct {
	ID        ID
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Store persists sessions. Successful appends to one session form a total
// order and are returned by Load in that order. Different sessions may be
// accessed concurrently.
type Store interface {
	Create(context.Context, Metadata) (ID, error)
	Append(context.Context, ID, Entry) error
	Load(context.Context, ID) ([]Entry, error)
	List(context.Context, Query) ([]Summary, error)
}

// MutableStore extends Store with mutable session metadata. Rename changes
// only the display title: it does not change the session identity, entry
// sequence, CreatedAt, or UpdatedAt. Rename calls for one session are ordered
// with that session's Append and Load operations.
type MutableStore interface {
	Store
	Rename(context.Context, ID, string) error
}

// ErrNotFound indicates that a session does not exist.
var ErrNotFound = errors.New("session not found")
