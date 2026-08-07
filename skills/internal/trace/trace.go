package trace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type contextKey struct{}

type Logger struct {
	mu      sync.Mutex
	writer  io.Writer
	enabled bool
}

func NewLogger(writer io.Writer, enabled bool) *Logger {
	if writer == nil {
		writer = os.Stdout
	}
	return &Logger{writer: writer, enabled: enabled}
}

func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

func FromContext(ctx context.Context) *Logger {
	logger, _ := ctx.Value(contextKey{}).(*Logger)
	return logger
}

func Log(ctx context.Context, function string, fields map[string]any) {
	if logger := FromContext(ctx); logger != nil {
		logger.Log(function, fields)
	}
}

func (l *Logger) Log(function string, fields map[string]any) {
	if l == nil || !l.enabled {
		return
	}
	if fields == nil {
		fields = map[string]any{}
	}

	entry := map[string]any{
		"location": callerLocation(),
		"ts":       time.Now().Format(time.RFC3339Nano),
		"function": function,
	}
	for key, value := range fields {
		entry[key] = sanitize(key, value)
	}

	data, err := marshalEntry(entry)
	if err != nil {
		data = []byte(fmt.Sprintf(`{"ts":%q,"function":%q,"error":%q}`, time.Now().Format(time.RFC3339Nano), function, err.Error()))
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.writer, "[trace] %s\n", data)
}

func marshalEntry(entry map[string]any) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteByte('{')
	first := true
	for _, key := range orderedKeys(entry) {
		keyData, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		valueData, err := json.Marshal(entry[key])
		if err != nil {
			return nil, err
		}
		if !first {
			buffer.WriteByte(',')
		}
		first = false
		buffer.Write(keyData)
		buffer.WriteByte(':')
		buffer.Write(valueData)
	}
	buffer.WriteByte('}')
	return buffer.Bytes(), nil
}

func orderedKeys(entry map[string]any) []string {
	keys := make([]string, 0, len(entry))
	for key := range entry {
		if key != "location" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return append([]string{"location"}, keys...)
}

func callerLocation() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown:0"
	}
	return fmt.Sprintf("%s:%d", projectRelativePath(file), line)
}

func projectRelativePath(file string) string {
	file = filepath.ToSlash(file)
	marker := "/ai-study/"
	if idx := strings.LastIndex(file, marker); idx >= 0 {
		return file[idx+len(marker):]
	}
	return filepath.Base(file)
}

func sanitize(key string, value any) any {
	lower := strings.ToLower(key)
	if strings.Contains(lower, "api_key") && !strings.HasSuffix(lower, "_set") {
		return "***"
	}
	if strings.Contains(lower, "authorization") || strings.Contains(lower, "secret") || strings.Contains(lower, "access_token") || strings.Contains(lower, "refresh_token") {
		return "***"
	}
	return value
}
