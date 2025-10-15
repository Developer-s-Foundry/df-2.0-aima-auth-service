package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func (p *PostgresConn) GetUser(context context.Context, email string) (*User, error) {
	query := `
		SELECT userId, username, email, hashedPassword, roleid, created_at, updated_at, expires_at
		FROM users
		WHERE email = $1
	`

	t := &User{}

	err := p.Conn.QueryRow(
		context, query, email,
	).Scan(&t.UserID,
		&t.UserName,
		&t.Email,
		&t.HashedPassword,
		&t.RoleId,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with email: %s not found", email)
		}
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}
	return t, nil
}
