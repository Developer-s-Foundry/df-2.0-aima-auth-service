package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func (p *PostgresConn) GetUser(context context.Context, email string) (*User, error) {
	query := `
		SELECT userId, username, email, hashedPassword, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	t := &User{}

	err := p.Conn.QueryRow(
		context, query, email,
	).Scan(&t.UserID,
		&t.Email,
		&t.HashedPassword,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with email: %s not found", email)
		}
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}
	return t, nil
}
