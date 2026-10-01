package platform

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"

	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorCyan   = "\033[36m"
	colorBlue   = "\033[34m"
	colorWhite  = "\033[37m"
	colorGray   = "\033[90m"

	bgGreen  = "\033[42m"
	bgYellow = "\033[43m"
	bgRed    = "\033[41m"
	bgBlue   = "\033[44m"
	bgCyan   = "\033[46m"
)

// statusColor returns an ANSI color for an HTTP status code.
// Green=2xx, Yellow=3xx/4xx, Red=5xx
func statusColor(code int) string {
	switch {
	case code >= 500:
		return colorRed + colorBold
	case code >= 400:
		return colorYellow + colorBold
	case code >= 300:
		return colorCyan
	default:
		return colorGreen + colorBold
	}
}

// methodColor returns a color for an HTTP method badge.
func methodColor(method string) string {
	switch method {
	case http.MethodGet:
		return colorBlue + colorBold
	case http.MethodPost:
		return colorGreen + colorBold
	case http.MethodDelete:
		return colorRed + colorBold
	case http.MethodPut, http.MethodPatch:
		return colorYellow + colorBold
	default:
		return colorWhite + colorBold
	}
}

// durationColor colors the request duration based on latency thresholds.
func durationColor(d time.Duration) string {
	switch {
	case d > 500*time.Millisecond:
		return colorRed
	case d > 100*time.Millisecond:
		return colorYellow
	default:
		return colorGreen
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// Hijack implements http.Hijacker so WebSocket upgrades work through this middleware.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
	}
	return h.Hijack()
}

// Flush implements http.Flusher so SSE streaming works through this middleware.
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// RequestLogger is an HTTP middleware that logs every request with colored output.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		ts := time.Now().Format("15:04:05.000")

		isWS := r.Header.Get("Upgrade") == "websocket"

		tsColor := colorGreen + colorBold
		if rw.statusCode >= 500 {
			tsColor = colorRed + colorBold
		} else if rw.statusCode >= 400 {
			tsColor = colorYellow + colorBold
		}

		if isWS {
			fmt.Printf(
				"%s%s%s  %s%-9s%s  %s%-40s%s  %sWS%s\n",
				tsColor, ts, colorReset,
				colorCyan+colorBold, "WS", colorReset,
				colorWhite, r.URL.Path, colorReset,
				colorCyan+colorBold, colorReset,
			)
			return
		}

		fmt.Printf(
			"%s%s%s  %s%-7s%s  %s%-40s%s  %s%3d%s  %s%s%s  %s%s%s\n",
			tsColor, ts, colorReset,
			methodColor(r.Method), r.Method, colorReset,
			colorGray, r.URL.RequestURI(), colorReset,
			statusColor(rw.statusCode), rw.statusCode, colorReset,
			durationColor(duration), formatDuration(duration), colorReset,
			colorGray, r.RemoteAddr, colorReset,
		)
	})
}

// LogInfo prints a structured info log line in green.
func LogInfo(component, msg string, kv ...any) {
	ts := time.Now().Format("15:04:05.000")
	kvStr := formatKV(kv...)
	fmt.Printf("%s%s%s  %sINF%s  [%s%s%s] %s%s%s%s\n",
		colorGreen+colorBold, ts, colorReset,
		colorGreen+colorBold, colorReset,
		colorCyan, component, colorReset,
		colorWhite, msg, colorReset,
		kvStr,
	)
}

// LogWarn prints a structured warning log line in yellow.
func LogWarn(component, msg string, kv ...any) {
	ts := time.Now().Format("15:04:05.000")
	kvStr := formatKV(kv...)
	fmt.Printf("%s%s%s  %sWRN%s  [%s%s%s] %s%s%s%s\n",
		colorYellow+colorBold, ts, colorReset,
		colorYellow+colorBold, colorReset,
		colorCyan, component, colorReset,
		colorWhite, msg, colorReset,
		kvStr,
	)
}

// LogError prints a structured error log line in red.
func LogError(component, msg string, kv ...any) {
	ts := time.Now().Format("15:04:05.000")
	kvStr := formatKV(kv...)
	fmt.Printf("%s%s%s  %sERR%s  [%s%s%s] %s%s%s%s\n",
		colorRed+colorBold, ts, colorReset,
		colorRed+colorBold, colorReset,
		colorCyan, component, colorReset,
		colorWhite+colorBold, msg, colorReset,
		kvStr,
	)
}

// LogStartup prints a banner on server start.
func LogStartup(gatewayID, port, dbPath string) {
	fmt.Println()
	fmt.Printf("  %s┌─────────────────────────────────────────┐%s\n", colorCyan, colorReset)
	fmt.Printf("  %s│%s  %s⚡  WhatsApp-Style Messaging Gateway%s       %s│%s\n", colorCyan, colorReset, colorWhite+colorBold, colorReset, colorCyan, colorReset)
	fmt.Printf("  %s├─────────────────────────────────────────┤%s\n", colorCyan, colorReset)
	fmt.Printf("  %s│%s  Gateway  %-30s %s│%s\n", colorCyan, colorReset, colorGreen+colorBold+gatewayID+colorReset, colorCyan, colorReset)
	fmt.Printf("  %s│%s  Port     %-30s %s│%s\n", colorCyan, colorReset, colorGreen+colorBold+":"+port+colorReset, colorCyan, colorReset)
	fmt.Printf("  %s│%s  Database %-30s %s│%s\n", colorCyan, colorReset, colorGray+dbPath+colorReset, colorCyan, colorReset)
	fmt.Printf("  %s│%s  Episode  %s1 – Core Mechanics%s              %s│%s\n", colorCyan, colorReset, colorYellow+colorBold, colorReset, colorCyan, colorReset)
	fmt.Printf("  %s└─────────────────────────────────────────┘%s\n", colorCyan, colorReset)
	fmt.Println()
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dμs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func formatKV(kv ...any) string {
	if len(kv) == 0 {
		return ""
	}
	out := "  "
	for i := 0; i+1 < len(kv); i += 2 {
		out += fmt.Sprintf("%s%v%s=%s%v%s ", colorGray, kv[i], colorReset, colorWhite, kv[i+1], colorReset)
	}
	return out
}
