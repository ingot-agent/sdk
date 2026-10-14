package model_test

import (
	"bytes"
	"sync"
	"testing"

	"github.com/ingot-agent/sdk/model"
)

func TestReplayStateOwnership(t *testing.T) {
	var absent *model.ReplayState
	if absent.Clone() != nil {
		t.Fatal("nil state became present")
	}
	original := &model.ReplayState{Format: "example.signed-output.v1", Provider: "p", Model: "m", Data: []byte{0, 255, 1}}
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			copy := original.Clone()
			if copy == original || copy.Format != original.Format || copy.Provider != original.Provider || copy.Model != original.Model || !bytes.Equal(copy.Data, original.Data) {
				t.Error("state not copied")
			}
			copy.Data[0] = 42
		}()
	}
	group.Wait()
	if original.Data[0] != 0 {
		t.Fatal("copy aliases original")
	}
	for _, data := range [][]byte{nil, {}} {
		copy := (&model.ReplayState{Data: data}).Clone()
		if (copy.Data == nil) != (data == nil) {
			t.Fatal("nil/empty data changed")
		}
	}
}
