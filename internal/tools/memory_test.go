package tools

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zieckey/ai-study/internal/memory"
)

func TestMemoryToolSetGetListDelete(t *testing.T) {
	tool := Memory{Store: memory.NewStore(filepath.Join(t.TempDir(), "memory.json"))}
	ctx := context.Background()

	setInput, _ := json.Marshal(map[string]string{"action": "set", "key": "language", "value": "Go"})
	if _, err := tool.Execute(ctx, setInput); err != nil {
		t.Fatalf("set returned error: %v", err)
	}

	getInput, _ := json.Marshal(map[string]string{"action": "get", "key": "language"})
	got, err := tool.Execute(ctx, getInput)
	if err != nil {
		t.Fatalf("get returned error: %v", err)
	}
	if !strings.Contains(got, "Go") {
		t.Fatalf("get result = %q", got)
	}

	listInput, _ := json.Marshal(map[string]string{"action": "list"})
	got, err = tool.Execute(ctx, listInput)
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}
	if !strings.Contains(got, "language") {
		t.Fatalf("list result = %q", got)
	}

	deleteInput, _ := json.Marshal(map[string]string{"action": "delete", "key": "language"})
	if _, err := tool.Execute(ctx, deleteInput); err != nil {
		t.Fatalf("delete returned error: %v", err)
	}
}
