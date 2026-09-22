package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// handler adapts slog to bgscan's line format and fans lines out to
// live-view subscribers. Line format:
// "2006/01/02 15:04:05 [LEVEL] message key=value ...".
type handler struct {
	sink   *sink
	attrs  []slog.Attr // accumulated via With
	groups []string    // accumulated via WithGroup
	// Enabled is answered from the shared level dial.
}

var _ slog.Handler = (*handler)(nil)

func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	return h.sink != nil && level >= h.sink.level.Level()
}

func (h *handler) Handle(_ context.Context, r slog.Record) error {
	if h.sink == nil || h.sink.file == nil {
		return nil
	}

	var b strings.Builder
	b.WriteString(r.Time.Format("2006/01/02 15:04:05"))
	b.WriteString(" [")
	b.WriteString(r.Level.String())
	b.WriteString("] ")
	b.WriteString(r.Message)

	for _, a := range h.attrs {
		writeAttr(&b, h.groups, a)
	}
	r.Attrs(func(a slog.Attr) bool {
		writeAttr(&b, h.groups, a)
		return true
	})

	line := b.String()

	_, _ = io.WriteString(h.sink.file, line+"\n")
	h.sink.broadcast(line)

	return nil
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	next := &handler{
		sink:   h.sink,
		attrs:  make([]slog.Attr, 0, len(h.attrs)+len(attrs)),
		groups: h.groups,
	}
	next.attrs = append(next.attrs, h.attrs...)
	next.attrs = append(next.attrs, attrs...)
	return next
}

func (h *handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	next := &handler{
		sink:   h.sink,
		attrs:  h.attrs,
		groups: make([]string, 0, len(h.groups)+1),
	}
	next.groups = append(next.groups, h.groups...)
	next.groups = append(next.groups, name)
	return next
}

// writeAttr appends " key=value", nesting inside active groups.
func writeAttr(b *strings.Builder, groups []string, a slog.Attr) {
	if a.Equal(slog.Attr{}) {
		return
	}
	b.WriteByte(' ')
	for _, g := range groups {
		b.WriteString(g)
		b.WriteByte('.')
	}
	b.WriteString(a.Key)
	b.WriteByte('=')
	b.WriteString(attrValue(a.Value))
}

// attrValue renders a value with minimal quoting so lines stay greppable.
func attrValue(v slog.Value) string {
	if v.Kind() == slog.KindString {
		s := v.String()
		if s == "" || strings.ContainsAny(s, " \t\n\"=") {
			return strconv.Quote(s)
		}
		return s
	}
	if v.Kind() == slog.KindTime {
		return v.Time().Format(time.RFC3339)
	}
	return fmt.Sprint(v.Any())
}
