package postgres

import (
	"context"
	"fmt"
	"time"
)

func (p *PostgresConn) InsertUser(u User) error {
	query := `
		INSERT INTO users (userId, email, hashedPassword)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := p.Conn.QueryRow(
		ctx, query, u.UserID,
		u.Email, u.HashedPassword,
	).Scan(&u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}
