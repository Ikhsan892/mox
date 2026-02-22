package logs

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"mox/tools/stack"
)

// Log handler for all application
// example :
//
// logger := slog.New(logs.NewBaseLogHandler(&logs.LogOptions{
// 	AddSource: false,
// 	BatchSize: 10,
// 	MinLevel:  slog.LevelInfo,
// 	Filterrable: func(ctx context.Context, log *logs.Log) bool {
// 	},
// 	WriteFunc: func(ctx context.Context, log []*logs.Log) error {
// 		for _, log := range log {
// 			fmt.Println(log.Time, log.Message)
// 		}
// 		return nil
// 	},
// }))

type Log struct {
	Time    time.Time
	Message string
	Level   slog.Level
	Data    map[string]any
	Source  string
	Payload []byte
}

var _ slog.Handler = (*LogHandler)(nil)

type LogOptions struct {
	// add source information to log
	AddSource bool

	// filter logs function
	// filter logs that should be displayed or not
	Filterrable func(ctx context.Context, logByte []byte, log *Log) bool

	WriteFunc func(ctx context.Context, log []*Log) error
	// MinLevel specifies what level should be write
	// for filtering logs
	MinLevel slog.Leveler
	// BatchSize specifies how many logs to accumulate before calling WriteFunc.
	// If not set or 0, fallback to 100 by default.
	BatchSize int
}

type LogHandler struct {
	mu      *sync.Mutex
	options *LogOptions
	buf     []byte
	logs    []*Log
	groups  []groupOrAttr
}

func NewBaseLogHandler(opt *LogOptions) *LogHandler {
	l := &LogHandler{
		mu:      &sync.Mutex{},
		options: opt,
	}

	l.buf = make([]byte, 1024)

	if opt.WriteFunc == nil {
		panic("WriteFunc must be set")
	}

	if l.options.MinLevel == nil {
		l.options.MinLevel = slog.LevelInfo
	}

	if l.options.BatchSize == 0 {
		l.options.BatchSize = 100
	}

	return l
}

// Enabled implements slog.Handler.
func (l *LogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level.Level() >= l.options.MinLevel.Level()
}

type groupOrAttr struct {
	group string
	attrs []slog.Attr
}

func (l *LogHandler) withGroupAttrs(param groupOrAttr) *LogHandler {
	l2 := *l
	l2.groups = make([]groupOrAttr, len(l.groups)+1)
	copy(l2.groups, l.groups)
	l2.groups = append(l2.groups, param)
	l2.groups[len(l2.groups)-1] = param

	return &l2
}

// Handle implements slog.Handler.
func (l *LogHandler) Handle(ctx context.Context, record slog.Record) error {
	goas := l.groups
	if record.NumAttrs() == 0 {
		// If the record has no Attrs, remove groups at the end of the list; they are empty.
		for len(goas) > 0 && goas[len(goas)-1].group != "" {
			goas = goas[:len(goas)-1]
		}
	}

	data := make(map[string]any, record.NumAttrs())
	stack := stack.New[slog.Attr]()

	if len(goas) > 0 {
		for i := len(goas) - 1; i >= 0; i-- {
			if goas[i].group != "" {
				l.buf = fmt.Appendf(l.buf, "%s.", goas[i].group)

				key := make(map[string]any, stack.Len())
				for _, a := range stack.PopByLength(stack.Len()) {
					buf, _ := l.appendAttr(key, a)
					l.buf = append(l.buf, string(buf)...)
				}

				data[goas[i].group] = key
			} else if len(goas[i].attrs) > 0 {
				for _, a := range goas[i].attrs {
					stack.Push(a)
				}
			}
		}

		// ungroupped
		if stack.Len() > 0 {
			for _, a := range stack.PopByLength(stack.Len()) {
				buf, _ := l.appendAttr(data, a)
				l.buf = append(l.buf, string(buf)...)
			}
		}
	}

	source := ""

	if record.PC != 0 && l.options.AddSource {
		fs := runtime.CallersFrames([]uintptr{record.PC})
		f, _ := fs.Next()
		l.appendAttr(data, slog.String(slog.SourceKey, fmt.Sprintf("%s:%d", f.File, f.Line)))
		source = fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
	}

	record.Attrs(func(a slog.Attr) bool {
		buf, err := l.appendAttr(data, a)
		l.buf = append(l.buf, string(buf)...)

		if err != nil {
			return false
		}
		return true
	})

	log := &Log{
		Time:    record.Time,
		Message: record.Message,
		Level:   record.Level,
		Data:    data,
		Source:  source,
		Payload: l.buf,
	}

	if l.options.Filterrable != nil && !l.options.Filterrable(ctx, l.buf, log) {
		return nil
	}

	l.mu.Lock()
	l.logs = append(l.logs, log)
	l.buf = nil
	logLength := len(l.logs)
	l.mu.Unlock()

	if logLength >= l.options.BatchSize && l.options.WriteFunc != nil {
		if err := l.WriteAll(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (l *LogHandler) WriteAll(ctx context.Context) error {

	l.mu.Lock()

	totalLogs := len(l.logs)

	// no logs to write
	if totalLogs == 0 {
		l.mu.Unlock()
		return nil
	}

	logs := make([]*Log, totalLogs)
	copy(logs, l.logs)
	l.logs = l.logs[:0]

	l.mu.Unlock()

	return l.options.WriteFunc(ctx, logs)
}

// WithAttrs implements slog.Handler.
func (l *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return l
	}

	return l.withGroupAttrs(groupOrAttr{attrs: attrs})
}

// WithGroup implements slog.Handler.
func (l *LogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return l
	}

	return l.withGroupAttrs(groupOrAttr{group: name})
}

func (l *LogHandler) appendAttr(data map[string]any, attr slog.Attr) ([]byte, error) {
	attr.Value = attr.Value.Resolve()
	buf := make([]byte, 1024)

	// skip empty attributes
	if attr.Equal(slog.Attr{}) {
		return []byte{}, nil
	}

	switch attr.Value.Kind() {
	case slog.KindGroup:
		attrs := attr.Value.Group()

		if len(attrs) == 0 {
			return []byte{}, nil
		}

		groupData := make(map[string]any, len(attrs))
		for _, a := range attrs {
			b, _ := l.appendAttr(groupData, a)
			buf = fmt.Appendf(buf, "%s.%s", []byte(attr.Key), b)
		}

		if len(groupData) > 0 {
			data[attr.Key] = groupData
		}
	case slog.KindInt64:
		v := attr.Value.Int64()
		buf = fmt.Appendf(buf, "%s=%d ", attr.Key, v)
		data[attr.Key] = v

	case slog.KindUint64:
		v := attr.Value.Uint64()
		buf = fmt.Appendf(buf, "%s=%d", attr.Key, v)
		data[attr.Key] = v
	case slog.KindString:
		v := attr.Value.String()
		buf = fmt.Appendf(buf, "%s=%s ", attr.Key, v)
		data[attr.Key] = v

	default:
		v := attr.Value.Any()

		if err, ok := v.(error); ok {
			buf = fmt.Appendf(buf, "%s=%s ", attr.Key, err.Error())
			data[attr.Key] = err.Error()
		} else {
			buf = fmt.Appendf(buf, "%s=%s ", attr.Key, v)
			data[attr.Key] = v
		}
	}

	return buf, nil
}
