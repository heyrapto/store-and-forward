package sqlite

import (
	"context"
	"database/sql"
	"time"

	"messaging/server/internal/domain"
)

type DeliveryRepo struct {
	DB *sql.DB
}

func NewDeliveryRepo(db *sql.DB) *DeliveryRepo {
	return &DeliveryRepo{DB: db}
}

func (r *DeliveryRepo) Save(ctx context.Context, d domain.Delivery) error {
	query := `INSERT INTO deliveries (message_id, recipient_id, seq, status, expires_at) 
	          VALUES (?, ?, ?, ?, ?) ON CONFLICT DO NOTHING`
	_, err := r.DB.ExecContext(ctx, query, 
		d.MessageID, d.RecipientID, d.Seq, d.Status, d.ExpiresAt.Format(time.RFC3339))
	return err
}

func (r *DeliveryRepo) FindPending(ctx context.Context, recipientID domain.DeviceID, afterSeq uint64) ([]domain.Delivery, error) {
	query := `
		SELECT message_id, recipient_id, seq, status, expires_at 
		FROM deliveries 
		WHERE recipient_id = ? AND seq > ? AND status = 'pending' 
		ORDER BY seq ASC`
	
	rows, err := r.DB.QueryContext(ctx, query, recipientID, afterSeq)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []domain.Delivery
	for rows.Next() {
		var d domain.Delivery
		var expires string
		if err := rows.Scan(&d.MessageID, &d.RecipientID, &d.Seq, &d.Status, &expires); err != nil {
			return nil, err
		}
		d.ExpiresAt, _ = time.Parse(time.RFC3339, expires)
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}

// PendingMessage holds a delivery row joined with its message payload.
type PendingMessage struct {
	domain.Delivery
	ConversationID string
	SenderID       domain.DeviceID
	Payload        string
	SentAt         time.Time
	ServerAt       time.Time
}

// FindPendingWithMessages returns pending deliveries joined with their full message payloads.
func (r *DeliveryRepo) FindPendingWithMessages(ctx context.Context, recipientID domain.DeviceID, afterSeq uint64) ([]PendingMessage, error) {
	query := `
		SELECT d.message_id, d.recipient_id, d.seq, d.status, d.expires_at,
		       m.conversation_id, m.sender_id, m.payload, m.sent_at, m.server_at
		FROM deliveries d
		JOIN messages m ON m.id = d.message_id
		WHERE d.recipient_id = ? AND d.seq > ? AND d.status = 'pending'
		ORDER BY d.seq ASC`

	rows, err := r.DB.QueryContext(ctx, query, recipientID, afterSeq)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PendingMessage
	for rows.Next() {
		var pm PendingMessage
		var expires, sentAt, serverAt string
		if err := rows.Scan(
			&pm.MessageID, &pm.RecipientID, &pm.Seq, &pm.Status, &expires,
			&pm.ConversationID, &pm.SenderID, &pm.Payload, &sentAt, &serverAt,
		); err != nil {
			return nil, err
		}
		pm.ExpiresAt, _ = time.Parse(time.RFC3339, expires)
		pm.SentAt, _ = time.Parse(time.RFC3339, sentAt)
		pm.ServerAt, _ = time.Parse(time.RFC3339, serverAt)
		results = append(results, pm)
	}
	return results, rows.Err()
}

func (r *DeliveryRepo) Advance(ctx context.Context, messageID domain.MessageID, recipientID domain.DeviceID, next domain.DeliveryStatus) error {
	now := time.Now().UTC().Format(time.RFC3339)
	var query string
	
	if next == domain.DeliveryDelivered {
		query = `UPDATE deliveries SET status = ?, delivered_at = ? 
		         WHERE message_id = ? AND recipient_id = ? AND status = 'pending'`
	} else if next == domain.DeliveryRead {
		query = `UPDATE deliveries SET status = ?, read_at = ? 
		         WHERE message_id = ? AND recipient_id = ? AND (status = 'pending' OR status = 'delivered')`
	} else {
		return domain.ErrInvalidTransition
	}

	res, err := r.DB.ExecContext(ctx, query, next, now, messageID, recipientID)
	if err != nil {
		return err
	}
	
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		// Either already advanced or doesn't exist.
		// For idempotency, we return ErrAlreadyDelivered if we assume it's already past.
		// A full implementation would check the current status.
		return domain.ErrAlreadyDelivered 
	}
	return nil
}

func (r *DeliveryRepo) DeleteExpired(ctx context.Context) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	query := `DELETE FROM deliveries WHERE expires_at < ?`
	res, err := r.DB.ExecContext(ctx, query, now)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *DeliveryRepo) NextSeq(ctx context.Context, recipientID domain.DeviceID) (uint64, error) {
	// Atomic increment using a transaction. SQLite handles this serially.
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Insert or ignore if row doesn't exist
	_, err = tx.ExecContext(ctx, `INSERT INTO seq_counters (device_id, next_seq) VALUES (?, 1) ON CONFLICT DO NOTHING`, recipientID)
	if err != nil {
		return 0, err
	}

	var seq uint64
	err = tx.QueryRowContext(ctx, `SELECT next_seq FROM seq_counters WHERE device_id = ?`, recipientID).Scan(&seq)
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(ctx, `UPDATE seq_counters SET next_seq = next_seq + 1 WHERE device_id = ?`, recipientID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return seq, nil
}

func (r *DeliveryRepo) MailboxDepth(ctx context.Context, recipientID domain.DeviceID) (int64, error) {
	var count int64
	err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM deliveries WHERE recipient_id = ? AND status = 'pending'`, recipientID).Scan(&count)
	return count, err
}
