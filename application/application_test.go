package application_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/ingot-agent/sdk/application"
)

type process struct {
	mu        sync.Mutex
	arguments []string
	check     bool
	shutdowns []error
}

func (p *process) Arguments() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.arguments...)
}

func (p *process) Check() bool { return p.check }

func (p *process) Shutdown(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.shutdowns = append(p.shutdowns, err)
}

func TestProcessContextRoundTrip(t *testing.T) {
	want := &process{arguments: []string{"chat", "--plain"}, check: true}
	ctx := application.WithProcess(context.Background(), want)
	got, err := application.FromContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != want || !got.Check() {
		t.Fatalf("process=%#v check=%v", got, got.Check())
	}
	arguments := got.Arguments()
	arguments[0] = "changed"
	if got.Arguments()[0] != "chat" {
		t.Fatal("Arguments returned aliased storage")
	}
}

func TestProcessContextRejectsMissingAndNil(t *testing.T) {
	if _, err := application.FromContext(nil); !errors.Is(err, application.ErrUnavailable) {
		t.Fatalf("nil Context error=%v", err)
	}
	if _, err := application.FromContext(context.Background()); !errors.Is(err, application.ErrUnavailable) {
		t.Fatalf("missing Process error=%v", err)
	}
	assertPanics(t, func() { application.WithProcess(nil, &process{}) })
	var nilProcess *process
	assertPanics(t, func() { application.WithProcess(context.Background(), nilProcess) })
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("call did not panic")
		}
	}()
	fn()
}
