package pipeline_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ingot-agent/sdk/pipeline"
)

type recordingInterceptor struct {
	name  string
	trace *[]string
}

func (i recordingInterceptor) Invoke(
	ctx context.Context,
	req string,
	next pipeline.Next[string, string],
) (string, error) {
	*i.trace = append(*i.trace, i.name+" before")
	response, err := next(ctx, req+" "+i.name)
	*i.trace = append(*i.trace, i.name+" after")
	return response, err
}

func TestComposeUsesFirstInterceptorAsOutermost(t *testing.T) {
	t.Parallel()

	var trace []string
	terminal := func(_ context.Context, req string) (string, error) {
		trace = append(trace, "terminal")
		return req, nil
	}
	chain := pipeline.Compose(
		terminal,
		recordingInterceptor{name: "a", trace: &trace},
		recordingInterceptor{name: "b", trace: &trace},
		recordingInterceptor{name: "c", trace: &trace},
	)

	got, err := chain(context.Background(), "request")
	if err != nil {
		t.Fatal(err)
	}
	if got != "request a b c" {
		t.Fatalf("response = %q", got)
	}
	wantTrace := []string{
		"a before",
		"b before",
		"c before",
		"terminal",
		"c after",
		"b after",
		"a after",
	}
	if !reflect.DeepEqual(trace, wantTrace) {
		t.Fatalf("trace = %#v, want %#v", trace, wantTrace)
	}
}

type shortCircuitInterceptor struct {
	err error
}

func (i shortCircuitInterceptor) Invoke(
	context.Context,
	int,
	pipeline.Next[int, int],
) (int, error) {
	return 0, i.err
}

func TestComposeAllowsShortCircuit(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("stop")
	terminalCalled := false
	chain := pipeline.Compose(
		func(context.Context, int) (int, error) {
			terminalCalled = true
			return 1, nil
		},
		shortCircuitInterceptor{err: wantErr},
	)

	_, err := chain(context.Background(), 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v", err)
	}
	if terminalCalled {
		t.Fatal("short-circuited chain called terminal")
	}
}

func TestComposeWithoutInterceptorsReturnsTerminal(t *testing.T) {
	t.Parallel()

	terminal := pipeline.Next[int, int](func(_ context.Context, value int) (int, error) {
		return value + 1, nil
	})
	chain := pipeline.Compose(terminal)

	got, err := chain(context.Background(), 41)
	if err != nil || got != 42 {
		t.Fatalf("chain() = %d, %v", got, err)
	}
}

type contextKey struct{}

type contextCheckingInterceptor struct {
	want string
}

func (i contextCheckingInterceptor) Invoke(
	ctx context.Context,
	req int,
	next pipeline.Next[int, int],
) (int, error) {
	if got, _ := ctx.Value(contextKey{}).(string); got != i.want {
		return 0, errors.New("interceptor received a different context")
	}
	return next(ctx, req)
}

func TestComposePreservesContextAuthority(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), contextKey{}, "authority")
	chain := pipeline.Compose(
		func(gotCtx context.Context, value int) (int, error) {
			if gotCtx != ctx {
				return 0, errors.New("terminal received a different context")
			}
			return value, nil
		},
		contextCheckingInterceptor{want: "authority"},
	)

	got, err := chain(ctx, 42)
	if err != nil || got != 42 {
		t.Fatalf("chain() = %d, %v", got, err)
	}
}
