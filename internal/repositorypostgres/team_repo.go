package repositorypostgres

import (
	"context"
	"database/sql"
	"errors"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/repository"
)

type teamRepo struct {
	exec executor
}

func NewTeamRepository(db *DB) repository.TeamRepository {
	return &teamRepo{exec: db.SQL}
}

func (r *teamRepo) UpsertTeam(ctx context.Context, team domain.Team) (domain.Team, error) {
	if _, err := r.exec.ExecContext(ctx, `
		INSERT INTO teams (name) VALUES ($1)
		ON CONFLICT (name) DO NOTHING
	`, team.Name); err != nil {
		return domain.Team{}, err
	}

	var members []domain.User
	for _, m := range team.Members {
		row := r.exec.QueryRowContext(ctx, `
			INSERT INTO users (id, username, team_name, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE
			SET username = EXCLUDED.username,
			    team_name = EXCLUDED.team_name,
			    is_active = EXCLUDED.is_active
			RETURNING id, username, team_name, is_active
		`, m.ID, m.Username, team.Name, m.IsActive)

		var member domain.User
		if err := row.Scan(&member.ID, &member.Username, &member.TeamName, &member.IsActive); err != nil {
			return domain.Team{}, err
		}
		members = append(members, member)
	}

	return domain.Team{
		Name:    team.Name,
		Members: members,
	}, nil
}

func (r *teamRepo) GetTeamByName(ctx context.Context, teamName string) (domain.Team, error) {
	row := r.exec.QueryRowContext(ctx, `
		SELECT name FROM teams WHERE name = $1
	`, teamName)

	var team domain.Team
	if err := row.Scan(&team.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Team{}, domain.NewDomainError(domain.ErrorCodeNotFound, "team not found")
		}
		return domain.Team{}, err
	}

	rows, err := r.exec.QueryContext(ctx, `
		SELECT id, username, team_name, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY id
	`, teamName)
	if err != nil {
		return domain.Team{}, err
	}
	defer closeRows(rows)

	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return domain.Team{}, err
		}
		team.Members = append(team.Members, u)
	}

	return team, nil
}
