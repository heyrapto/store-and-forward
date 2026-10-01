package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"messaging/server/internal/domain"
)

type MessageRepo struct {
	DB *sql.DB
}

func NewMessageRepo(db *sql.DB) *MessageRepo {
	return &MessageRepo{DB: db}
}

func (r *MessageRepo) Save(ctx context.Context, m domain.Message) error {
	query := `INSERT INTO messages (id, conversation_id, sender_id, payload, encoding, sent_at, server_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.DB.ExecContext(ctx, query, 
		m.ID, m.ConversationID, m.SenderID, m.Payload, m.Encoding, 
		m.SentAt.Format(time.RFC3339), m.ServerAt.Format(time.RFC3339))
	
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "Duplicate") {
			return fmt.Errorf("%w: %s", domain.ErrDuplicateMessage, err.Error())
		}
		return err
	}
	return nil
}

func (r *MessageRepo) FindByID(ctx context.Context, id domain.MessageID) (domain.Message, error) {
	query := `SELECT id, conversation_id, sender_id, payload, encoding, sent_at, server_at 
	          FROM messages WHERE id = ?`
	row := r.DB.QueryRowContext(ctx, query, id)

	var m domain.Message
	var sent, server string
	err := row.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Payload, &m.Encoding, &sent, &server)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Message{}, domain.ErrMessageNotFound
		}
		return domain.Message{}, err
	}

	m.SentAt, _ = time.Parse(time.RFC3339, sent)
	m.ServerAt, _ = time.Parse(time.RFC3339, server)

	return m, nil
}
