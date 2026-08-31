package operation_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ingot-agent/sdk/interaction"
	"github.com/ingot-agent/sdk/operation"
	"github.com/ingot-agent/sdk/session"
)

type echoOperation struct{}

func (echoOperation) Definition() operation.Definition {
	return operation.Definition{
		Name:         "example.echo",
		Description:  "Return the supplied value.",
		InputSchema:  json.RawMessage(`{"type":"object","additionalProperties":false,"required":["value"],"properties":{"value":{"type":"string"}}}`),
		OutputSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"required":["value"],"properties":{"value":{"type":"string"}}}`),
	}
}

func (echoOperation) Invoke(ctx context.Context, request operation.Request) (operation.Result, error) {
	if err := ctx.Err(); err != nil {
		return operation.Result{}, err
	}
	return operation.Result{Output: append(json.RawMessage(nil), request.Input...)}, nil
}

var _ operation.Operation = echoOperation{}

func TestExternalOperationContract(t *testing.T) {
	candidate := echoOperation{}
	definition := candidate.Definition()
	definition.InputSchema[0] = 'x'
	if candidate.Definition().InputSchema[0] != '{' {
		t.Fatal("Definition retained caller mutation")
	}

	input := json.RawMessage(`{"value":"hello"}`)
	result, err := candidate.Invoke(context.Background(), operation.Request{
		SessionID:   session.ID("session-1"),
		Input:       input,
		Interaction: interaction.Unavailable(),
	})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	input[0] = 'x'
	if string(result.Output) != `{"value":"hello"}` {
		t.Fatalf("Output = %s", result.Output)
	}
	result.Output[0] = 'y'
	if input[0] != 'x' {
		t.Fatal("Result retained request input")
	}
}
