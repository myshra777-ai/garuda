// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrorLogEntry is one row's worth of slog data, captured for async
// persistence to errors_log.
type ErrorLogEntry struct {
	Time      time.Time
	Level     string
	Message   string
	RequestID string
	Attrs     map[string]any
}

// ErrorLogHandler wraps an slog.Handler. Every record is forwarded to
// the underlying handler. Warn and Error records are also enqueued for
// async persistence.
//
// The channel is buffered. When full, entries are dropped and a
// counter is incremented. The handler never blocks the caller — a
// slow database must not stall request handling.
type ErrorLogHandler struct {
	inner   slog.Handler
	entries chan ErrorLogEntry
	dropped *atomic.Int64
}

// NewErrorLogHandler wraps inner with an async error-log sink.
// bufferSize is the channel capacity. 1024 is enough for a single
// server under normal load; drop is expected only during a sustained
// database outage.
func NewErrorLogHandler(inner slog.Handler, bufferSize int) *ErrorLogHandler {
	if bufferSize <= 0 {
		bufferSize = 1024
	}
	return &ErrorLogHandler{
		inner:   inner,
		entries: make(chan ErrorLogEntry, bufferSize),
		dropped: &atomic.Int64{},
	}
}

func (h *ErrorLogHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.inner.Enabled(ctx, l)
}

func (h *ErrorLogHandler) Handle(ctx context.Context, r slog.Record) error {
	// Always forward to the underlying handler.
	_ = h.inner.Handle(ctx, r)

	// Only Warn and Error are persisted.
	if r.Level < slog.LevelWarn {
		return nil
	}

	entry := ErrorLogEntry{
		Time:    r.Time,
		Level:   r.Level.String(),
		Message: r.Message,
		Attrs:   recordAttrs(r),
	}

	// request_id may be in attrs (the standard pattern in this
	// codebase) or in the context. Check attrs first, then context.
	if v, ok := entry.Attrs["request_id"]; ok {
		if s, ok := v.(string); ok {
			entry.RequestID = s
		}
	}
	if entry.RequestID == "" && ctx != nil {
		if s, ok := ctx.Value("request_id").(string); ok {
			entry.RequestID = s
		}
	}

	select {
	case h.entries <- entry:
	default:
		h.dropped.Add(1)
	}
	return nil
}

func (h *ErrorLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ErrorLogHandler{
		inner:   h.inner.WithAttrs(attrs),
		entries: h.entries,
		dropped: h.dropped,
	}
}

func (h *ErrorLogHandler) WithGroup(name string) slog.Handler {
	return &ErrorLogHandler{
		inner:   h.inner.WithGroup(name),
		entries: h.entries,
		dropped: h.dropped,
	}
}

// Dropped returns the number of entries discarded because the buffer
// was full.
func (h *ErrorLogHandler) Dropped() int64 {
	return h.dropped.Load()
}

// StartErrorLogWriter drains the handler's entry channel into
// errors_log until ctx is cancelled. A write failure drops the entry
// (reported to stderr, never to slog — logging about logging would
// recurse).
func StartErrorLogWriter(ctx context.Context, h *ErrorLogHandler, pool *pgxpool.Pool) {
	if pool == nil || h == nil {
		return
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-h.entries:
				writeErrorEntry(pool, e)
			}
		}
	}()
}

func writeErrorEntry(pool *pgxpool.Pool, e ErrorLogEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	details, err := json.Marshal(e.Attrs)
	if err != nil {
		details = []byte("{}")
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO errors_log (occurred_at, level, message, request_id, details)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5::jsonb)
	`, e.Time, e.Level, e.Message, e.RequestID, string(details))
	if err != nil {
		fmt.Fprintf(os.Stderr, "errors_log write failed: %v\n", err)
	}
}

func recordAttrs(r slog.Record) map[string]any {
	attrs := make(map[string]any, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = attrValue(a.Value)
		return true
	})
	return attrs
}

func attrValue(v slog.Value) any {
	v = v.Resolve()
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindInt64:
		return v.Int64()
	case slog.KindUint64:
		return v.Uint64()
	case slog.KindFloat64:
		return v.Float64()
	case slog.KindBool:
		return v.Bool()
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindTime:
		return v.Time().Format(time.RFC3339)
	case slog.KindGroup:
		group := v.Group()
		m := make(map[string]any, len(group))
		for _, ga := range group {
			m[ga.Key] = attrValue(ga.Value)
		}
		return m
	default:
		return fmt.Sprintf("%v", v.Any())
	}
}
