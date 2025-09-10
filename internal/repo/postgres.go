package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Uva337/WBL0v1/internal/models"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context) (*Postgres, error) {
	host := envOr("PG_HOST", "localhost")
	port := envOr("PG_PORT", "5432")
	user := envOr("PG_USER", "order_user")
	pass := envOr("PG_PASSWORD", "order_pass")
	db := envOr("PG_DB", "order_db")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, db)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &Postgres{pool: pool}, nil
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) UpsertOrder(ctx context.Context, o models.Order) error {
	b, err := json.Marshal(o)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) 

	_, err = tx.Exec(ctx, `INSERT INTO orders(order_uid, payload) VALUES($1, $2)
        ON CONFLICT (order_uid) DO UPDATE SET payload = EXCLUDED.payload`, o.OrderUID, b)
	if err != nil {
		return fmt.Errorf("failed to execute upsert: %w", err)
	}

	return tx.Commit(ctx)
}

func (p *Postgres) GetOrder(ctx context.Context, id string) (models.Order, bool, error) {
	var raw []byte
	err := p.pool.QueryRow(ctx, `SELECT payload FROM orders WHERE order_uid=$1`, id).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Order{}, false, nil
		}
		return models.Order{}, false, fmt.Errorf("failed to query order: %w", err)
	}

	var o models.Order
	if err := json.Unmarshal(raw, &o); err != nil {
		return models.Order{}, false, fmt.Errorf("failed to unmarshal order payload: %w", err)
	}
	return o, true, nil
}

func (p *Postgres) GetAll(ctx context.Context) ([]models.Order, error) {
	rows, err := p.pool.Query(ctx, `SELECT payload FROM orders`)
	if err != nil {
		return nil, fmt.Errorf("failed to query all orders: %w", err)
	}
	defer rows.Close()

	var list []models.Order
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("failed to scan order row: %w", err)
		}
		var o models.Order
		if err := json.Unmarshal(raw, &o); err != nil {
			return nil, fmt.Errorf("failed to unmarshal order from db: %w", err)
		}
		list = append(list, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading rows: %w", err)
	}
	return list, nil
}
