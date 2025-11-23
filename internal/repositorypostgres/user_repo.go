package repositorypostgres

import (
	"context"
	"database/sql"
	"errors"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/repository"
)

type userRepo struct {
	exec executor
}

func NewUserRepository(db *DB) repository.UserRepository {
	return &userRepo{exec: db.SQL}
}

func (r *userRepo) GetUserByID(ctx context.Context, userID string) (domain.User, error) {
	row := r.exec.QueryRowContext(ctx, `
		SELECT id, username, team_name, is_active
		FROM users
		WHERE id = $1
	`, userID)

	var u domain.User
	if err := row.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.NewDomainError(domain.ErrorCodeNotFound, "user not found")
		}
		return domain.User{}, err
	}
	return u, nil
}

func (r *userRepo) SetActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	row := r.exec.QueryRowContext(ctx, `
		UPDATE users
		SET is_active = $2
		WHERE id = $1
		RETURNING id, username, team_name, is_active
	`, userID, isActive)

	var u domain.User
	if err := row.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.NewDomainError(domain.ErrorCodeNotFound, "user not found")
		}
		return domain.User{}, err
	}
	return u, nil
}

func (r *userRepo) ListActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	rows, err := r.exec.QueryContext(ctx, `
		SELECT id, username, team_name, is_active
		FROM users
		WHERE team_name = $1 AND is_active = TRUE
		ORDER BY id
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
