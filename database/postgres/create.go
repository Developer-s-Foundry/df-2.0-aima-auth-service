package postgres

import (
	"context"
	"time"
)

type User struct {
	UserID         string    `json:"userId"`
	UserName       string    `json:"username"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"hashedPassword"`
	RoleId         string    `json:"roleId"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
	DeletedAt      time.Time `json:"deleted_at,omitempty"`
}

func (p *PostgresConn) Create() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			userId TEXT PRIMARY KEY,
			username VARCHAR(100),
			email VARCHAR(100) NOT NULL,
			hashedPassword TEXT NOT NULL,
			roleId VARCHAR(100),
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
	)	
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := p.Conn.Exec(ctx, query)
	return err
}
