package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const profileStoreTimeout = 5 * time.Second

type ProfileStore interface {
	CreateProfile(params CreateProfileParams) (User, error)
	FindProfileByID(id string) (User, error)
}

type CreateProfileParams struct {
	ID              string
	Email           string
	DisplayName     string
	Username        string
	EmailVisibility string
}

type PostgresProfileStore struct {
	pool *pgxpool.Pool
}

func NewPostgresProfileStore(pool *pgxpool.Pool) *PostgresProfileStore {
	return &PostgresProfileStore{pool: pool}
}

func (s *PostgresProfileStore) CreateProfile(params CreateProfileParams) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), profileStoreTimeout)
	defer cancel()

	const query = `
		insert into public.users (id, display_name, username, email_visibility)
		values ($1, $2, $3, $4)
		returning id, display_name, username, coalesce(profile_text, ''), coalesce(avatar_url, ''), email_visibility, created_at, updated_at
	`

	user := User{
		Email: params.Email,
	}

	err := s.pool.QueryRow(
		ctx,
		query,
		params.ID,
		strings.TrimSpace(params.DisplayName),
		normalizeUsername(params.Username),
		normalizeEmailVisibility(params.EmailVisibility),
	).Scan(
		&user.ID,
		&user.DisplayName,
		&user.Username,
		&user.ProfileText,
		&user.AvatarURL,
		&user.EmailVisibility,
		&user.CreatedAt,
		new(time.Time),
	)
	if err != nil {
		return User{}, mapProfileStoreError(err)
	}

	return user, nil
}

func (s *PostgresProfileStore) FindProfileByID(id string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), profileStoreTimeout)
	defer cancel()

	const query = `
		select id, display_name, username, coalesce(profile_text, ''), coalesce(avatar_url, ''), email_visibility, created_at
		from public.users
		where id = $1
	`

	var user User
	err := s.pool.QueryRow(ctx, query, strings.TrimSpace(id)).Scan(
		&user.ID,
		&user.DisplayName,
		&user.Username,
		&user.ProfileText,
		&user.AvatarURL,
		&user.EmailVisibility,
		&user.CreatedAt,
	)
	if err != nil {
		return User{}, mapProfileStoreError(err)
	}

	return user, nil
}

func mapProfileStoreError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserProfileNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_username_key":
				return ErrUsernameAlreadyExists
			}
		}
	}

	return fmt.Errorf("profile store: %w", err)
}
