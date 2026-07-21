package lots

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	selectByID = `SELECT d, title, description, start_price, current_price, current_winner_id, 
       status, starts_at, ends_at, photo_path FROM lots WHERE id = $1`
	selectAll  = `SELECT id, title, description, start_price, current_price, status, starts_at, 
       ends_at, photo_path FROM lots WHERE status = 'active' ORDER BY id`
)

type LotRepo interface {
	CreateLot(ctx context.Context, title, description string, startPrice float64, photo string,
		endsAt time.Time, status string, sellerID int64, currentPrice float64) error
	GetAll(ctx context.Context) ([]model.Lots, error)
}
type lots struct {
	lots *pgxpool.Pool
}

func NewLots(pool *pgxpool.Pool) (LotRepo, error) {
	return &lots{lots: pool}, nil
}

func (l *lots) CreateLot(ctx context.Context, title, description string, startPrice float64,
	photo string, endsAt time.Time, status string, sellerID int64, currentPrice float64) error {
	_, err := l.lots.Exec(ctx, `INSERT INTO lots (title, description, start_price, photo_path, 
                  ends_at, status, seller_id, current_price) 
     VALUES ($1,$2, $3,$4, $5, $6, $7, $8)`,
		title, description, startPrice, photo, endsAt, status, sellerID, currentPrice)
	if err != nil {
		return model.ErrDatabase
	}
	return nil
}

func (l *lots) GetAll(ctx context.Context) ([]model.Lots, error) {
	rows, err := l.lots.Query(ctx, selectAll)
	if err != nil {
		return nil, fmt.Errorf("get lots: %w", err)
	}
	defer rows.Close()

	var out []model.Lots
	for rows.Next() {
		var p model.Lots
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.StartPrice, &p.CurrentPrice,
			&p.Status, &p.StartAt, &p.EndAt, &p.Photo); err != nil {
			return nil, fmt.Errorf("get lots: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get lots: %w", err)
	}
	return out, nil
}

// UpdateLot в процессе.
func (l *lots) UpdateLot(ctx context.Context, id int64, status string, currentPrice int64) error {
	err := l.lots.QueryRow(ctx, `UPDATE lots SET status = $1, current_price = $2 WHERE id = $3`,
		status, currentPrice, id).Scan(&id, &status, &currentPrice)
	if err != nil {
		return model.ErrDatabase
	}
	return nil
}

func (l *lots) DeleteLots(ctx context.Context, id int64) {}
