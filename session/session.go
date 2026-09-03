// Package session defines opaque session persistence, lifecycle management,
// and discovery capabilities.
package session

import (
	"context"
	"errors"
	"time"
)

// ID is the stable identity of one session. Store implementations create IDs;
// callers must not assume that an arbitrary ID is a valid persistent identity.
type ID string

// Metadata describes the authoritative lifecycle state of one session.
// CreatedAt is the successful creation time. UpdatedAt is the time of the last
// committed Entry mutation and is not changed by lifecycle operations. An
// empty session has UpdatedAt equal to CreatedAt. ArchivedAt is nil for an
// active session and otherwise records the successful archive time.
type Metadata struct {
	ID         ID
	Title      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt *time.Time
}

// Entry is an opaque, durable, versioned persistence record. Version identifies
// the payload schema for Kind. Session implementations preserve Kind, Version,
// Payload bytes, and append order without interpreting the payload. Aggregate
// inputs are immutable by contract and returned payloads are caller-owned.
type Entry struct {
	Kind    string
	Version int
	Payload []byte
}

// CreateRequest describes caller-controlled properties of a new session. The
// Store creates its ID and timestamps.
type CreateRequest struct {
	Title string
}

// ForkRequest describes caller-controlled properties of a fork target. An
// empty Title copies the source title.
type ForkRequest struct {
	Title string
}

// Store provides the persistence primitives for opaque session entries.
// Implementations are safe for concurrent use. Successful appends to one
// session form an order and Load returns committed entries in that order.
//
// Create returns authoritative metadata for a newly created, active, empty
// session. The implementation owns the ID, CreatedAt, and UpdatedAt values.
// Load is permitted for active and archived sessions and returns caller-owned
// entries.
//
// Append to an archived session fails with ErrArchived. Append returning nil
// means the entry is definitely committed and UpdatedAt reflects its commit
// time. Append returning an error means commit status is unknown to the caller;
// callers must not retry merely because an error was returned. The commit
// decision is final when Append returns: an Append that committed is visible
// to every successful Load begun after that return, and an Append that did not
// commit will not appear later.
type Store interface {
	Create(context.Context, CreateRequest) (Metadata, error)
	Append(context.Context, ID, Entry) error
	Load(context.Context, ID) ([]Entry, error)
}

// Manager manages the lifecycle of known session identities. Get and every
// mutation return authoritative metadata on success. Rename is permitted for
// active and archived sessions and changes only Title. Archive and Restore are
// idempotent desired-state operations and do not change UpdatedAt; repeated
// Archive preserves the original ArchivedAt. Delete removes either an active
// or archived session but does not imply forensic secure erase and must not
// delete assets referenced by opaque entries.
//
// Fork creates a new active session whose entries are an opaque logical copy
// of the source's committed entries in order. An archived source may be forked.
// The target has a new identity and creation timestamps and does not inherit
// the source lifecycle state. The portable contract does not define whether a
// concurrently appended source entry falls before or after the fork boundary,
// but the target must never contain a partial or corrupt entry. Callers needing
// a deterministic boundary must prevent concurrent source mutation.
//
// The portable contract likewise does not define one cross-implementation
// linearization order between Delete and concurrent Store or Manager calls.
type Manager interface {
	Get(context.Context, ID) (Metadata, error)
	Rename(context.Context, ID, string) (Metadata, error)
	Archive(context.Context, ID) (Metadata, error)
	Restore(context.Context, ID) (Metadata, error)
	Delete(context.Context, ID) error
	Fork(context.Context, ID, ForkRequest) (Metadata, error)
}

// Query discovers sessions. List returns caller-owned authoritative metadata
// for both active and archived sessions. Implementations should use a
// deterministic ordering and document it when callers may rely on that order.
type Query interface {
	List(context.Context) ([]Metadata, error)
}

var (
	// ErrNotFound indicates that the target session does not exist.
	ErrNotFound = errors.New("session not found")
	// ErrArchived indicates that an operation is invalid because the target
	// session is archived.
	ErrArchived = errors.New("session archived")
)
