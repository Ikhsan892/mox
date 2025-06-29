package logs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPanicLogHandler(t *testing.T) {

	defer func() {
		if err := recover(); err == nil {
			assert.Fail(t, "Expected to panic in log handler")
		}
	}()

	NewBaseLogHandler(&LogOptions{})

}

func TestNewLogHandler(t *testing.T) {
	l := NewBaseLogHandler(&LogOptions{
		AddSource: false,
		WriteFunc: func(_ context.Context, log []*Log) error {

			return nil
		},
	})

	assert.Equal(t, 100, l.options.BatchSize)
	assert.Nil(t, l.options.Filterrable)
	assert.Equal(t, slog.LevelInfo, l.options.MinLevel)
}

func TestLogHandlerEnabled(t *testing.T) {
	h := NewBaseLogHandler(&LogOptions{
		MinLevel: slog.LevelWarn,
		WriteFunc: func(_ context.Context, log []*Log) error {
			return nil
		},
	})

	l := slog.New(h)

	scenarios := []struct {
		level    slog.Level
		expected bool
	}{
		{slog.LevelDebug, false},
		{slog.LevelInfo, false},
		{slog.LevelWarn, true},
		{slog.LevelError, true},
	}

	for _, s := range scenarios {
		t.Run(fmt.Sprintf("Level %v", s.level), func(t *testing.T) {
			result := l.Enabled(context.Background(), s.level)

			if result != s.expected {
				t.Fatalf("Expected %v, got %v", s.expected, result)
			}
		})
	}
}

func TestLogHandlerWithAttrsAndWithGroup(t *testing.T) {
	h0 := NewBaseLogHandler(&LogOptions{
		WriteFunc: func(_ context.Context, log []*Log) error {
			return nil
		},
	})

	h1 := h0.WithAttrs([]slog.Attr{slog.Int("test1", 1)}).(*LogHandler)
	h2 := h1.WithGroup("h2_group").(*LogHandler)
	h3 := h2.WithAttrs([]slog.Attr{slog.Int("test2", 2)}).(*LogHandler)

	scenarios := []struct {
		name           string
		handler        *LogHandler
		expectedParent *LogHandler
		expectedGroup  string
		expectedAttrs  int
	}{
		{
			"h0",
			h0,
			nil,
			"",
			0,
		},
		{
			"h1",
			h1,
			h0,
			"",
			1,
		},
		{
			"h2",
			h2,
			h1,
			"",
			0,
		},
		{
			"h3",
			h3,
			h2,
			"h2_group",
			0,
		},
	}

	for i, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			if len(s.handler.groups) > 0 {
				if s.handler.groups[i].group != s.expectedGroup {
					t.Fatalf("Expected group %q, got %q", s.expectedGroup, s.handler.groups[i].group)
				}

				if totalAttrs := len(s.handler.groups[i].attrs); totalAttrs != s.expectedAttrs {
					t.Fatalf("Expected %d attrs, got %d", s.expectedAttrs, totalAttrs)
				}
			}
		})
	}
}

func TestLogHandlerHandle(t *testing.T) {
	ctx := context.Background()

	beforeLogs := []*Log{}
	writeLogs := []*Log{}

	h := NewBaseLogHandler(&LogOptions{
		BatchSize: 3,
		Filterrable: func(ctx context.Context, logByte []byte, log *Log) bool {
			beforeLogs = append(beforeLogs, log)

			if log.Message == "test2" {
				return false // skip test2 log
			}

			return true
		},
		WriteFunc: func(_ context.Context, log []*Log) error {
			writeLogs = log
			return nil
		},
	})

	h.Handle(ctx, slog.NewRecord(time.Now(), slog.LevelInfo, "test1", 0))
	h.Handle(ctx, slog.NewRecord(time.Now(), slog.LevelInfo, "test2", 0))
	h.Handle(ctx, slog.NewRecord(time.Now(), slog.LevelInfo, "test3", 0))

	// no batch write
	{
		checkLogMessages([]string{"test1", "test2", "test3"}, beforeLogs, t)

		checkLogMessages([]string{"test1", "test3"}, h.logs, t)

		// should be empty because no batch write has happened yet
		if totalWriteLogs := len(writeLogs); totalWriteLogs != 0 {
			t.Fatalf("Expected %d writeLogs, got %d", 0, totalWriteLogs)
		}
	}

	// add one more log to trigger the batch write
	{
		h.Handle(ctx, slog.NewRecord(time.Now(), slog.LevelInfo, "test4", 0))

		// should be empty after the batch write
		checkLogMessages([]string{}, h.logs, t)

		checkLogMessages([]string{"test1", "test3", "test4"}, writeLogs, t)
	}
}

func TestLogHandlerWriteAll(t *testing.T) {
	ctx := context.Background()

	beforeLogs := []*Log{}
	writeLogs := []*Log{}

	h := NewBaseLogHandler(&LogOptions{
		BatchSize: 3,
		Filterrable: func(ctx context.Context, logByte []byte, log *Log) bool {
			beforeLogs = append(beforeLogs, log)

			return true
		},
		WriteFunc: func(_ context.Context, log []*Log) error {
			writeLogs = log
			return nil
		},
	})

	h.Handle(ctx, slog.NewRecord(time.Now(), slog.LevelInfo, "test1", 0))
	h.Handle(ctx, slog.NewRecord(time.Now(), slog.LevelInfo, "test2", 0))

	checkLogMessages([]string{"test1", "test2"}, beforeLogs, t)
	checkLogMessages([]string{"test1", "test2"}, h.logs, t)
	checkLogMessages([]string{}, writeLogs, t) // empty because the batch size hasn't been reached

	// force trigger the batch write
	h.WriteAll(ctx)

	checkLogMessages([]string{"test1", "test2"}, beforeLogs, t)
	checkLogMessages([]string{}, h.logs, t) // reset
	checkLogMessages([]string{"test1", "test2"}, writeLogs, t)
}

func TestBatchHandlerAttrsFormat(t *testing.T) {
	ctx := context.Background()

	beforeLogs := []*Log{}

	h0 := NewBaseLogHandler(&LogOptions{
		Filterrable: func(_ context.Context, logByte []byte, log *Log) bool {
			beforeLogs = append(beforeLogs, log)
			return true
		},
		WriteFunc: func(_ context.Context, logs []*Log) error {
			return nil
		},
	})

	h1 := h0.WithAttrs([]slog.Attr{slog.Int("a", 1), slog.String("b", "123")})

	h2 := h1.WithGroup("sub").WithAttrs([]slog.Attr{
		slog.Int("c", 3),
		slog.Any("d", map[string]any{"d.1": 1}),
		slog.Any("e", errors.New("example error")),
	})

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "hello", 0)
	record.AddAttrs(slog.String("name", "test"))

	h0.Handle(ctx, record)
	h1.Handle(ctx, record)
	h2.Handle(ctx, record)

	expected := []string{
		`{"name":"test"}`,
		`{"a":1,"b":"123","name":"test"}`,
		`{"a":1,"b":"123","name":"test","sub":{"c":3,"d":{"d.1":1},"e":"example error"}}`,
	}

	if len(beforeLogs) != len(expected) {
		t.Fatalf("Expected %d logs, got %d", len(beforeLogs), len(expected))
	}

	for i, data := range expected {
		t.Run(fmt.Sprintf("log handler %d", i), func(t *testing.T) {
			log := beforeLogs[i]
			raw, _ := json.Marshal(log.Data)
			if string(raw) != data {
				t.Fatalf("Expected \n%s \ngot \n%s", data, raw)
			}
		})
	}
}

func checkLogMessages(expected []string, logs []*Log, t *testing.T) {
	if len(logs) != len(expected) {
		t.Fatalf("Expected %d batched logs, got %d (expected: %v)", len(expected), len(logs), expected)
	}

	for _, message := range expected {
		exists := false
		for _, l := range logs {
			if l.Message == message {
				exists = true
				continue
			}
		}
		if !exists {
			t.Fatalf("Missing %q log message", message)
		}
	}
}
