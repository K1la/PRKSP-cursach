package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/username/parking-service/internal/model"
)

type BookingRepository struct {
	db *pgxpool.Pool
}

type SpotBookingInfo struct {
	SpotID       uuid.UUID
	IsAvailable  bool
	ParkingLotID uuid.UUID
	OwnerID      uuid.UUID
	LotActive    bool
	PricePerHour float64
}

type CreateBookingParams struct {
	UserID        uuid.UUID
	ParkingSpotID uuid.UUID
	StartTime     time.Time
	EndTime       time.Time
	TotalPrice    float64
	VehiclePlate  string
}

func NewBookingRepository(db *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) GetSpotBookingInfo(ctx context.Context, spotID uuid.UUID) (SpotBookingInfo, error) {
	var info SpotBookingInfo
	err := r.db.QueryRow(ctx, `
		SELECT ps.id,
		       ps.is_available,
		       pl.id,
		       pl.owner_id,
		       pl.is_active,
		       pl.price_per_hour::float8
		FROM parking_spots ps
		JOIN parking_lots pl ON pl.id = ps.parking_lot_id
		WHERE ps.id = $1
	`, spotID).Scan(&info.SpotID, &info.IsAvailable, &info.ParkingLotID, &info.OwnerID, &info.LotActive, &info.PricePerHour)
	if errors.Is(err, pgx.ErrNoRows) {
		return SpotBookingInfo{}, ErrNotFound
	}
	return info, err
}

func (r *BookingRepository) HasOverlap(ctx context.Context, spotID uuid.UUID, start time.Time, end time.Time) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM bookings
			WHERE parking_spot_id = $1
			  AND status IN ('pending', 'active')
			  AND start_time < $3
			  AND end_time > $2
		)
	`, spotID, start, end).Scan(&exists)
	return exists, err
}

func (r *BookingRepository) Create(ctx context.Context, params CreateBookingParams) (model.Booking, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.Booking{}, err
	}
	defer tx.Rollback(ctx)

	var isAvailable bool
	var lotActive bool
	err = tx.QueryRow(ctx, `
		SELECT ps.is_available, pl.is_active
		FROM parking_spots ps
		JOIN parking_lots pl ON pl.id = ps.parking_lot_id
		WHERE ps.id = $1
		FOR UPDATE OF ps
	`, params.ParkingSpotID).Scan(&isAvailable, &lotActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Booking{}, ErrNotFound
	}
	if err != nil {
		return model.Booking{}, err
	}
	if !isAvailable || !lotActive {
		return model.Booking{}, ErrConflict
	}

	var overlap bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM bookings
			WHERE parking_spot_id = $1
			  AND status IN ('pending', 'active')
			  AND start_time < $3
			  AND end_time > $2
		)
	`, params.ParkingSpotID, params.StartTime, params.EndTime).Scan(&overlap)
	if err != nil {
		return model.Booking{}, err
	}
	if overlap {
		return model.Booking{}, ErrConflict
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO bookings (user_id, parking_spot_id, start_time, end_time, total_price, vehicle_plate)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, parking_spot_id, start_time, end_time, status, total_price::float8,
		          vehicle_plate, created_at, updated_at
	`, params.UserID, params.ParkingSpotID, params.StartTime, params.EndTime, params.TotalPrice, params.VehiclePlate)
	booking, err := scanBooking(row)
	if err != nil {
		return model.Booking{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Booking{}, err
	}
	return booking, nil
}

func (r *BookingRepository) GetByID(ctx context.Context, id uuid.UUID) (model.Booking, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, parking_spot_id, start_time, end_time, status, total_price::float8,
		       vehicle_plate, created_at, updated_at
		FROM bookings
		WHERE id = $1
	`, id)
	return scanBooking(row)
}

func (r *BookingRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]model.Booking, error) {
	return r.list(ctx, `
		SELECT id, user_id, parking_spot_id, start_time, end_time, status, total_price::float8,
		       vehicle_plate, created_at, updated_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
}

func (r *BookingRepository) ListByParkingLot(ctx context.Context, lotID uuid.UUID, limit int, offset int) ([]model.Booking, error) {
	return r.list(ctx, `
		SELECT b.id, b.user_id, b.parking_spot_id, b.start_time, b.end_time, b.status,
		       b.total_price::float8, b.vehicle_plate, b.created_at, b.updated_at
		FROM bookings b
		JOIN parking_spots ps ON ps.id = b.parking_spot_id
		WHERE ps.parking_lot_id = $1
		ORDER BY b.created_at DESC
		LIMIT $2 OFFSET $3
	`, lotID, limit, offset)
}

func (r *BookingRepository) Cancel(ctx context.Context, id uuid.UUID) (model.Booking, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE bookings
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND status IN ('pending', 'active')
		RETURNING id, user_id, parking_spot_id, start_time, end_time, status, total_price::float8,
		          vehicle_plate, created_at, updated_at
	`, id)
	return scanBooking(row)
}

func (r *BookingRepository) list(ctx context.Context, query string, args ...any) ([]model.Booking, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]model.Booking, 0)
	for rows.Next() {
		booking, err := scanBooking(rows)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	return bookings, rows.Err()
}

func scanBooking(row rowScanner) (model.Booking, error) {
	var booking model.Booking
	err := row.Scan(
		&booking.ID,
		&booking.UserID,
		&booking.ParkingSpotID,
		&booking.StartTime,
		&booking.EndTime,
		&booking.Status,
		&booking.TotalPrice,
		&booking.VehiclePlate,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Booking{}, ErrNotFound
	}
	return booking, err
}
