package platform

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
)

// ColorHandler is a custom slog.Handler that outputs colorized log lines.
// Compatible with the slog.Logger the WS handler already uses.
type ColorHandler struct {
	component string
	level     slog.Level
}

func (h *ColorHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *ColorHandler) Handle(_ context.Context, r slog.Record) error {
	ts := r.Time.Format("15:04:05.000")

	var lvlColor, lvlLabel string
	switch {
	case r.Level >= slog.LevelError:
		lvlColor = colorRed + colorBold
		lvlLabel = "ERR"
	case r.Level >= slog.LevelWarn:
		lvlColor = colorYellow + colorBold
		lvlLabel = "WRN"
	default:
		lvlColor = colorGreen + colorBold
		lvlLabel = "INF"
	}

	// Build KV pairs from attrs
	var kvBuf bytes.Buffer
	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(&kvBuf, "  %s%s%s=%s%v%s", colorGray, a.Key, colorReset, colorWhite, a.Value.Any(), colorReset)
		return true
	})

	fmt.Fprintf(os.Stdout, "%s%s%s  %s%s%s  [%s%s%s] %s%s%s%s\n",
		lvlColor, ts, colorReset,
		lvlColor, lvlLabel, colorReset,
		colorCyan, h.component, colorReset,
		colorWhite, r.Message, colorReset,
		kvBuf.String(),
	)
	return nil
}

func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h // simplified — attrs folded into each record in real impl
}

func (h *ColorHandler) WithGroup(name string) slog.Handler {
	return h
}

// NewWSLogger returns a *slog.Logger that outputs colorised lines using the
// platform ColorHandler. Pass it to the WS handler so all websocket events
// appear in the same visual style as the HTTP request log.
func NewWSLogger() *slog.Logger {
	return slog.New(&ColorHandler{component: "ws", level: slog.LevelDebug})
}
