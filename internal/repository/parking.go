package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/username/parking-service/internal/model"
)

type ParkingRepository struct {
	db *pgxpool.Pool
}

type ParkingLotFilter struct {
	Query     string
	Latitude  *float64
	Longitude *float64
	RadiusKM  *float64
	Limit     int
	Offset    int
}

type CreateParkingLotParams struct {
	OwnerID      uuid.UUID
	Name         string
	Description  *string
	Address      string
	Latitude     float64
	Longitude    float64
	TotalSpots   int
	PricePerHour float64
}

type UpdateParkingLotParams struct {
	Name         string
	Description  *string
	Address      string
	Latitude     float64
	Longitude    float64
	TotalSpots   int
	PricePerHour float64
	IsActive     bool
}

func NewParkingRepository(db *pgxpool.Pool) *ParkingRepository {
	return &ParkingRepository{db: db}
}

func (r *ParkingRepository) List(ctx context.Context, filter ParkingLotFilter) ([]model.ParkingLot, error) {
	query := `
		SELECT id, owner_id, name, description, address, latitude::float8, longitude::float8,
		       total_spots, price_per_hour::float8, is_active, created_at, updated_at
		FROM parking_lots
		WHERE is_active = TRUE
	`
	args := []any{}
	if strings.TrimSpace(filter.Query) != "" {
		args = append(args, "%"+strings.TrimSpace(filter.Query)+"%")
		query += fmt.Sprintf(` AND (name ILIKE $%d OR address ILIKE $%d)`, len(args), len(args))
	}
	if filter.Latitude != nil && filter.Longitude != nil && filter.RadiusKM != nil {
		args = append(args, *filter.Latitude, *filter.Longitude, *filter.RadiusKM)
		query += fmt.Sprintf(`
			AND (
				6371 * acos(
					LEAST(1, GREATEST(-1,
						cos(radians($%d)) * cos(radians(latitude::float8)) *
						cos(radians(longitude::float8) - radians($%d)) +
						sin(radians($%d)) * sin(radians(latitude::float8))
					))
				)
			) <= $%d
		`, len(args)-2, len(args)-1, len(args)-2, len(args))
	}
	args = append(args, filter.Limit, filter.Offset)
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, len(args)-1, len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lots := make([]model.ParkingLot, 0)
	for rows.Next() {
		lot, err := scanParkingLot(rows)
		if err != nil {
			return nil, err
		}
		lots = append(lots, lot)
	}
	return lots, rows.Err()
}

func (r *ParkingRepository) GetByID(ctx context.Context, id uuid.UUID) (model.ParkingLot, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, owner_id, name, description, address, latitude::float8, longitude::float8,
		       total_spots, price_per_hour::float8, is_active, created_at, updated_at
		FROM parking_lots
		WHERE id = $1
	`, id)
	return scanParkingLot(row)
}

func (r *ParkingRepository) Create(ctx context.Context, params CreateParkingLotParams) (model.ParkingLot, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.ParkingLot{}, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		INSERT INTO parking_lots (
			owner_id, name, description, address, latitude, longitude, total_spots, price_per_hour
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, owner_id, name, description, address, latitude::float8, longitude::float8,
		          total_spots, price_per_hour::float8, is_active, created_at, updated_at
	`, params.OwnerID, params.Name, params.Description, params.Address, params.Latitude, params.Longitude, params.TotalSpots, params.PricePerHour)

	lot, err := scanParkingLot(row)
	if err != nil {
		return model.ParkingLot{}, err
	}

	if err := createDefaultSpots(ctx, tx, lot.ID, lot.TotalSpots); err != nil {
		return model.ParkingLot{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.ParkingLot{}, err
	}

	return lot, nil
}

func (r *ParkingRepository) Update(ctx context.Context, id uuid.UUID, params UpdateParkingLotParams) (model.ParkingLot, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.ParkingLot{}, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		UPDATE parking_lots
		SET name = $2,
		    description = $3,
		    address = $4,
		    latitude = $5,
		    longitude = $6,
		    total_spots = $7,
		    price_per_hour = $8,
		    is_active = $9,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, owner_id, name, description, address, latitude::float8, longitude::float8,
		          total_spots, price_per_hour::float8, is_active, created_at, updated_at
	`, id, params.Name, params.Description, params.Address, params.Latitude, params.Longitude, params.TotalSpots, params.PricePerHour, params.IsActive)
	lot, err := scanParkingLot(row)
	if err != nil {
		return model.ParkingLot{}, err
	}

	if err := createDefaultSpots(ctx, tx, lot.ID, lot.TotalSpots); err != nil {
		return model.ParkingLot{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.ParkingLot{}, err
	}

	return lot, nil
}

func (r *ParkingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM parking_lots WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ParkingRepository) ListSpots(ctx context.Context, lotID uuid.UUID) ([]model.ParkingSpot, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, parking_lot_id, spot_number, spot_type, is_available, floor, created_at
		FROM parking_spots
		WHERE parking_lot_id = $1
		ORDER BY spot_number
	`, lotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	spots := make([]model.ParkingSpot, 0)
	for rows.Next() {
		spot, err := scanParkingSpot(rows)
		if err != nil {
			return nil, err
		}
		spots = append(spots, spot)
	}
	return spots, rows.Err()
}

type batchSender interface {
	SendBatch(context.Context, *pgx.Batch) pgx.BatchResults
}

func createDefaultSpots(ctx context.Context, sender batchSender, lotID uuid.UUID, total int) error {
	batch := &pgx.Batch{}
	for i := 1; i <= total; i++ {
		batch.Queue(`
			INSERT INTO parking_spots (parking_lot_id, spot_number, spot_type)
			VALUES ($1, $2, 'standard')
			ON CONFLICT (parking_lot_id, spot_number) DO NOTHING
		`, lotID, formatSpotNumber(i))
	}
	results := sender.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < total; i++ {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func scanParkingLot(row rowScanner) (model.ParkingLot, error) {
	var lot model.ParkingLot
	err := row.Scan(
		&lot.ID,
		&lot.OwnerID,
		&lot.Name,
		&lot.Description,
		&lot.Address,
		&lot.Latitude,
		&lot.Longitude,
		&lot.TotalSpots,
		&lot.PricePerHour,
		&lot.IsActive,
		&lot.CreatedAt,
		&lot.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ParkingLot{}, ErrNotFound
	}
	return lot, err
}

func scanParkingSpot(row rowScanner) (model.ParkingSpot, error) {
	var spot model.ParkingSpot
	err := row.Scan(
		&spot.ID,
		&spot.ParkingLotID,
		&spot.SpotNumber,
		&spot.SpotType,
		&spot.IsAvailable,
		&spot.Floor,
		&spot.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ParkingSpot{}, ErrNotFound
	}
	return spot, err
}

func formatSpotNumber(number int) string {
	return fmt.Sprintf("A%02d", number)
}
