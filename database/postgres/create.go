package postgres

import (
	"context"
	"time"
)

type User struct {
	UserID         string    `json:"userId"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"hashedPassword"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}

func (p *PostgresConn) Create() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			userId TEXT PRIMARY KEY,
			email VARCHAR(100) NOT NULL,
			hashedPassword TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
	)	
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := p.Conn.Exec(ctx, query)
	return err
}
