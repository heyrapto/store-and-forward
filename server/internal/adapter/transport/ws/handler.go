package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"messaging/server/internal/adapter/persistence/sqlite"
	"messaging/server/internal/domain"
	"messaging/server/internal/usecase"
	"messaging/server/internal/usecase/port"
)

type Handler struct {
	Connect    *usecase.ConnectDevice
	Disconnect *usecase.DisconnectDevice
	Send       *usecase.SendMessage
	Ack        *usecase.AckDelivery
	Read       *usecase.MarkRead
	Deliveries *sqlite.DeliveryRepo
	Clock      port.Clock
	Logger     *slog.Logger
}

func NewHandler(c *usecase.ConnectDevice, d *usecase.DisconnectDevice, s *usecase.SendMessage, a *usecase.AckDelivery, r *usecase.MarkRead, deliveries *sqlite.DeliveryRepo, cl port.Clock, l *slog.Logger) *Handler {
	return &Handler{
		Connect:    c,
		Disconnect: d,
		Send:       s,
		Ack:        a,
		Read:       r,
		Deliveries: deliveries,
		Clock:      cl,
		Logger:     l,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		h.Logger.Error("websocket accept failed", slog.Any("err", err))
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal error")

	ctx := context.Background()

	// ── Handshake: read hello frame ────────────────────────────────────────────
	typ, msgBytes, err := c.Read(ctx)
	if err != nil || typ != websocket.MessageText {
		h.Logger.Warn("expected hello frame", slog.Any("err", err))
		c.Close(websocket.StatusUnsupportedData, "expected text hello frame")
		return
	}

	var env Envelope
	if err := json.Unmarshal(msgBytes, &env); err != nil || env.Type != "hello" {
		h.Logger.Warn("invalid hello frame", slog.String("type", env.Type))
		c.Close(websocket.StatusUnsupportedData, "expected hello frame")
		return
	}

	var hello HelloPayload
	if err := json.Unmarshal(env.Payload, &hello); err != nil {
		h.Logger.Warn("malformed hello payload", slog.Any("err", err))
		c.Close(websocket.StatusUnsupportedData, "invalid hello payload")
		return
	}

	// ── Connect use case ───────────────────────────────────────────────────────
	sendCh := make(chan []byte, 256)
	out, err := h.Connect.Execute(ctx, usecase.ConnectDeviceInput{
		DeviceID: hello.DeviceID,
		LastSeq:  hello.LastSeq,
		SendCh:   sendCh,
	})
	if err != nil {
		h.Logger.Error("connect device failed", slog.String("device_id", string(hello.DeviceID)), slog.Any("err", err))
		c.Close(websocket.StatusPolicyViolation, "connection failed")
		return
	}

	h.Logger.Info("device connected",
		slog.String("device_id", string(hello.DeviceID)),
		slog.Uint64("last_seq", hello.LastSeq),
		slog.Int("pending", len(out.PendingDeliveries)),
	)

	// ── Send hello_ok ──────────────────────────────────────────────────────────
	helloOk, _ := json.Marshal(Envelope{
		Type: "hello_ok",
		ID:   "srv-" + time.Now().Format("20060102150405"),
		Ts:   h.Clock.Now().Format(time.RFC3339),
		Payload: marshalPayload(HelloOkPayload{
			ConnectionID: "conn-" + string(hello.DeviceID),
			ServerTime:   h.Clock.Now().Format(time.RFC3339),
		}),
	})
	if err := c.Write(ctx, websocket.MessageText, helloOk); err != nil {
		h.Logger.Warn("failed to write hello_ok", slog.Any("err", err))
		return
	}

	// ── Send sync_batch for pending deliveries ─────────────────────────────────
	if len(out.PendingDeliveries) > 0 {
		pending, err := h.Deliveries.FindPendingWithMessages(ctx, hello.DeviceID, hello.LastSeq)
		if err != nil {
			h.Logger.Error("sync_batch query failed", slog.Any("err", err))
		} else if len(pending) > 0 {
			msgs := make([]DeliverPayload, 0, len(pending))
			for _, pm := range pending {
				msgs = append(msgs, DeliverPayload{
					MessageID:      string(pm.MessageID),
					ConversationID: pm.ConversationID,
					From:           string(pm.SenderID),
					Text:           pm.Payload,
					SentAt:         pm.SentAt.Format(time.RFC3339),
					ServerAt:       pm.ServerAt.Format(time.RFC3339),
					Seq:            pm.Seq,
				})
			}
			syncBatch, _ := json.Marshal(Envelope{
				Type: "sync_batch",
				ID:   "srv-sync-" + time.Now().Format("20060102150405"),
				Ts:   h.Clock.Now().Format(time.RFC3339),
				Payload: marshalPayload(SyncBatchPayload{
					Messages: msgs,
					HasMore:  false,
				}),
			})
			c.Write(ctx, websocket.MessageText, syncBatch)
		}
	}

	// ── Start read / write pumps ───────────────────────────────────────────────
	errc := make(chan error, 2)
	go func() { errc <- h.writePump(ctx, c, sendCh) }()
	go func() { errc <- h.readPump(ctx, c, hello.DeviceID) }()

	<-errc // first pump to exit wins

	h.Logger.Info("device disconnected", slog.String("device_id", string(hello.DeviceID)))
	h.Disconnect.Execute(context.Background(), usecase.DisconnectDeviceInput{DeviceID: hello.DeviceID})
	c.Close(websocket.StatusNormalClosure, "bye")
}

func (h *Handler) writePump(ctx context.Context, c *websocket.Conn, sendCh <-chan []byte) error {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-sendCh:
			if !ok {
				return nil // registry closed channel (session replaced)
			}
			if err := c.Write(ctx, websocket.MessageText, msg); err != nil {
				return err
			}
		case <-ticker.C:
			ping, _ := json.Marshal(Envelope{
				Type:    "ping",
				ID:      fmt.Sprintf("ping-%d", time.Now().Unix()),
				Ts:      h.Clock.Now().Format(time.RFC3339),
				Payload: []byte(`{}`),
			})
			if err := c.Write(ctx, websocket.MessageText, ping); err != nil {
				return err
			}
		}
	}
}

func (h *Handler) readPump(ctx context.Context, c *websocket.Conn, deviceID domain.DeviceID) error {
	for {
		typ, msgBytes, err := c.Read(ctx)
		if err != nil {
			return err
		}
		if typ != websocket.MessageText {
			continue
		}

		var env Envelope
		if err := json.Unmarshal(msgBytes, &env); err != nil {
			h.Logger.Warn("unreadable frame", slog.String("device_id", string(deviceID)), slog.Any("err", err))
			continue
		}

		switch env.Type {
		case "send":
			var p SendPayload
			if err := json.Unmarshal(env.Payload, &p); err != nil {
				continue
			}
			out, err := h.Send.Execute(ctx, usecase.SendMessageInput{
				ClientID: domain.MessageID(env.ID),
				To:       p.To,
				SenderID: deviceID,
				Payload:  p.Text,
				SentAt:   time.Now(),
			})
			if err != nil {
				h.Logger.Error("send message failed",
					slog.String("device_id", string(deviceID)),
					slog.String("to", p.To),
					slog.Any("err", err),
				)
				continue
			}
			h.Logger.Info("message sent",
				slog.String("from", string(deviceID)),
				slog.String("to", p.To),
				slog.String("msg_id", string(out.MessageID)),
			)
			ack, _ := json.Marshal(Envelope{
				Type: "ack_server",
				ID:   "srv-ack-" + string(out.MessageID),
				Ts:   out.ServerAt.Format(time.RFC3339),
				Payload: marshalPayload(AckServerPayload{
					RefID: string(out.MessageID),
				}),
			})
			c.Write(ctx, websocket.MessageText, ack)

		case "ack_device":
			var p AckDevicePayload
			if err := json.Unmarshal(env.Payload, &p); err == nil {
				if err := h.Ack.Execute(ctx, usecase.AckDeliveryInput{
					MessageID:   p.MessageID,
					RecipientID: deviceID,
				}); err != nil {
					h.Logger.Warn("ack_device failed",
						slog.String("msg_id", string(p.MessageID)),
						slog.Any("err", err),
					)
				}
			}

		case "read":
			var p ReadPayload
			if err := json.Unmarshal(env.Payload, &p); err == nil {
				h.Read.Execute(ctx, usecase.MarkReadInput{
					MessageID:   p.MessageID,
					RecipientID: deviceID,
				})
			}

		case "pong":
			// heartbeat reply, nothing to do

		default:
			h.Logger.Warn("unknown frame type",
				slog.String("type", env.Type),
				slog.String("device_id", string(deviceID)),
			)
		}
	}
}

func marshalPayload(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
