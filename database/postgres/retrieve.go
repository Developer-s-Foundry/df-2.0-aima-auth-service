package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (p *PostgresConn) GetUser(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT userId, email, hashedPassword, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	u := &User{}
	err := p.Conn.QueryRow(ctx, query, email).Scan(
		&u.UserID,
		&u.Email,
		&u.HashedPassword,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		fmt.Println("=------->")
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	return u, nil
}
