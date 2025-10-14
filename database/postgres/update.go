package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrUsernameTaken = errors.New("username is already taken")

func (p *PostgresConn) UpdateUsername(userID, newUsername string) error {
	query := `
		UPDATE users
		SET
			email      = $1,
			updated_at = $2
		WHERE user_id = $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := p.Conn.Exec(
		ctx,
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
