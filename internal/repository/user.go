package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/username/parking-service/internal/model"
)

type UserRepository struct {
	db *pgxpool.Pool
}

type CreateUserParams struct {
	Email        string
	PasswordHash string
	Name         string
	Phone        *string
	Role         model.UserRole
}

type UpdateUserParams struct {
	Name  string
	Phone *string
	Role  model.UserRole
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, params CreateUserParams) (model.User, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, name, phone, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, password_hash, name, phone, role, created_at, updated_at
	`, params.Email, params.PasswordHash, params.Name, params.Phone, params.Role)

	user, err := scanUser(row)
	if isUniqueViolation(err) {
		return model.User{}, ErrConflict
	}
	return user, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, name, phone, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email)
	return scanUser(row)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, name, phone, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id)
	return scanUser(row)
}

func (r *UserRepository) List(ctx context.Context, limit int, offset int) ([]model.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, email, password_hash, name, phone, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, params UpdateUserParams) (model.User, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE users
		SET name = $2, phone = $3, role = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, password_hash, name, phone, role, created_at, updated_at
	`, id, params.Name, params.Phone, params.Role)
	return scanUser(row)
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (model.User, error) {
	var user model.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.Phone,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	return user, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
