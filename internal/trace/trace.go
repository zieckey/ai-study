package trace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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
		writer = os.Stderr
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

	file, line := callerLocation()
	entry := map[string]any{
		"ts":       time.Now().Format(time.RFC3339Nano),
		"function": function,
		"file":     file,
		"line":     line,
	}
	for key, value := range fields {
		entry[key] = sanitize(key, value)
	}

	data, err := json.Marshal(entry)
	if err != nil {
		data = []byte(fmt.Sprintf(`{"ts":%q,"function":%q,"error":%q}`, time.Now().Format(time.RFC3339Nano), function, err.Error()))
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.writer, "[trace] %s\n", data)
}

func callerLocation() (string, int) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown", 0
	}
	return filepath.ToSlash(file), line
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
