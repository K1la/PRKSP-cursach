package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/username/parking-service/internal/model"
)

type ReviewRepository struct {
	db *pgxpool.Pool
}

type CreateReviewParams struct {
	UserID       uuid.UUID
	ParkingLotID uuid.UUID
	Rating       int
	Comment      *string
}

func NewReviewRepository(db *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(ctx context.Context, params CreateReviewParams) (model.Review, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO reviews (user_id, parking_lot_id, rating, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, parking_lot_id, rating, comment, created_at
	`, params.UserID, params.ParkingLotID, params.Rating, params.Comment)

	review, err := scanReview(row)
	if isUniqueViolation(err) {
		return model.Review{}, ErrConflict
	}
	return review, err
}

func (r *ReviewRepository) ListByParkingLot(ctx context.Context, lotID uuid.UUID, limit int, offset int) ([]model.Review, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, parking_lot_id, rating, comment, created_at
		FROM reviews
		WHERE parking_lot_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, lotID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]model.Review, 0)
	for rows.Next() {
		review, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, rows.Err()
}

func scanReview(row rowScanner) (model.Review, error) {
	var review model.Review
	err := row.Scan(
		&review.ID,
		&review.UserID,
		&review.ParkingLotID,
		&review.Rating,
		&review.Comment,
		&review.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Review{}, ErrNotFound
	}
	return review, err
}
