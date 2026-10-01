package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"messaging/server/internal/domain"
)

type GroupRepo struct {
	DB *sql.DB
}

func NewGroupRepo(db *sql.DB) *GroupRepo {
	return &GroupRepo{DB: db}
}

func (r *GroupRepo) FindByID(ctx context.Context, id domain.GroupID) (domain.Group, error) {
	query := `SELECT id, name, created_at FROM groups WHERE id = ?`
	row := r.DB.QueryRowContext(ctx, query, id)

	var g domain.Group
	var created string
	err := row.Scan(&g.ID, &g.Name, &created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Group{}, domain.ErrGroupNotFound
		}
		return domain.Group{}, err
	}

	g.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return g, nil
}

func (r *GroupRepo) Members(ctx context.Context, id domain.GroupID) ([]domain.DeviceID, error) {
	query := `SELECT device_id FROM group_members WHERE group_id = ?`
	rows, err := r.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.DeviceID
	for rows.Next() {
		var d domain.DeviceID
		if err := rows.Scan(&d); err != nil {
			return nil, err
		}
		members = append(members, d)
	}
	return members, rows.Err()
}

func (r *GroupRepo) AddMember(ctx context.Context, groupID domain.GroupID, deviceID domain.DeviceID) error {
	query := `INSERT INTO group_members (group_id, device_id) VALUES (?, ?) ON CONFLICT DO NOTHING`
	_, err := r.DB.ExecContext(ctx, query, groupID, deviceID)
	return err
}
