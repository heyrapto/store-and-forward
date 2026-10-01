package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"messaging/server/internal/adapter/bus/memory"
	"messaging/server/internal/adapter/persistence/sqlite"
	presencemem "messaging/server/internal/adapter/presence/memory"
	"messaging/server/internal/adapter/registry"
	"messaging/server/internal/adapter/transport/admin"
	rest "messaging/server/internal/adapter/transport/http"
	"messaging/server/internal/adapter/transport/ws"
	"messaging/server/internal/platform"
	"messaging/server/internal/usecase"
	"messaging/server/internal/usecase/port"
)

func main() {
	cfg := platform.LoadConfig()
	platform.LogStartup(cfg.GatewayID, cfg.Port, cfg.DBPath)

	// ── Storage & Adapters ─────────────────────────────────────────────────────
	db, err := sqlite.Open(cfg.DBPath, "./migrations")
	if err != nil {
		platform.LogError("db", "Failed to open database", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	platform.LogInfo("db", "Database ready", "path", cfg.DBPath)

	deviceRepo   := sqlite.NewDeviceRepo(db)
	groupRepo    := sqlite.NewGroupRepo(db)
	msgRepo      := sqlite.NewMessageRepo(db)
	deliveryRepo := sqlite.NewDeliveryRepo(db)

	presenceStore := presencemem.NewStore()
	eventBus      := memory.NewBus()
	connRegistry  := registry.NewRegistry()

	clk   := port.RealClock{}
	idGen := port.ULIDGenerator{}

	platform.LogInfo("adapters", "In-memory presence, bus, registry ready")

	// ── Use Cases ──────────────────────────────────────────────────────────────
	registerDevice := &usecase.RegisterDevice{
		Devices: deviceRepo,
		Groups:  groupRepo,
		Bus:     eventBus,
		Clock:   clk,
		IDs:     idGen,
	}

	connectDevice := &usecase.ConnectDevice{
		Devices:    deviceRepo,
		Deliveries: deliveryRepo,
		Presence:   presenceStore,
		Registry:   connRegistry,
		Bus:        eventBus,
		Clock:      clk,
		GatewayID:  cfg.GatewayID,
	}

	disconnectDevice := &usecase.DisconnectDevice{
		Presence: presenceStore,
		Registry: connRegistry,
		Bus:      eventBus,
		Clock:    clk,
	}

	sendMessage := &usecase.SendMessage{
		Messages:   msgRepo,
		Deliveries: deliveryRepo,
		Devices:    deviceRepo,
		Groups:     groupRepo,
		Presence:   presenceStore,
		Registry:   connRegistry,
		Bus:        eventBus,
		Clock:      clk,
		IDs:        idGen,
		TTL:        60 * time.Second,
	}

	ackDelivery := &usecase.AckDelivery{
		Messages:   msgRepo,
		Deliveries: deliveryRepo,
		Registry:   connRegistry,
		Bus:        eventBus,
		Clock:      clk,
	}

	markRead := &usecase.MarkRead{
		Messages:   msgRepo,
		Deliveries: deliveryRepo,
		Registry:   connRegistry,
		Bus:        eventBus,
		Clock:      clk,
	}

	sweepExpired := &usecase.SweepExpired{
		Deliveries: deliveryRepo,
	}

	// ── Background Jobs ────────────────────────────────────────────────────────
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			n, err := sweepExpired.Execute(context.Background())
			if err != nil {
				platform.LogWarn("sweeper", "Sweep failed", "err", err)
			} else if n > 0 {
				platform.LogInfo("sweeper", "Expired deliveries removed", "count", n)
			}
		}
	}()

	// ── Routes ─────────────────────────────────────────────────────────────────
	mux := http.NewServeMux()

	restHandler := rest.NewHandler(registerDevice, deviceRepo)
	restHandler.RegisterRoutes(mux)

	wsLogger := platform.NewWSLogger()
	wsHandler := ws.NewHandler(connectDevice, disconnectDevice, sendMessage, ackDelivery, markRead, deliveryRepo, clk, wsLogger)
	mux.Handle("/ws", wsHandler)

	sseHandler := admin.NewSSEHandler(eventBus)
	mux.Handle("/admin/events", sseHandler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","gateway_id":"%s","connections":%d}`, cfg.GatewayID, connRegistry.Count())
	})

	// ── Middleware Stack: CORS → Request Logger → Mux ─────────────────────────
	handler := platform.RequestLogger(
		corsMiddleware(mux),
	)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	// ── Start ──────────────────────────────────────────────────────────────────
	go func() {
		platform.LogInfo("http", "Listening", "addr", ":"+cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			platform.LogError("http", "Server error", "err", err)
		}
	}()

	// ── Graceful Shutdown ──────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	platform.LogWarn("http", "Shutting down…")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		platform.LogError("http", "Forced shutdown", "err", err)
	} else {
		platform.LogInfo("http", "Server stopped cleanly")
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
