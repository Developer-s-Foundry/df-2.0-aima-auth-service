package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrUsernameTaken = errors.New("username is already taken")

func (p *PostgresConn) UpdateUsername(context context.Context, userID, newUsername string) error {
	query := `
		UPDATE users
		SET
			email      = $1,
			updated_at = $2
		WHERE user_id = $3
	`

	result, err := p.Conn.Exec(
		context,
		query,
		newUsername,
		time.Now().UTC(),
		userID, // we are usign userId for updates
	)

	if err != nil {
		return fmt.Errorf("failed to update username: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with ID %s to update", userID)
	}

	return nil
}
