package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepository struct {
	db *pgxpool.Pool
}

type Stats struct {
	TotalUsers       int     `json:"total_users"`
	TotalBookings    int     `json:"total_bookings"`
	TotalParkingLots int     `json:"total_parking_lots"`
	Revenue          float64 `json:"revenue"`
}

func NewAdminRepository(db *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) Stats(ctx context.Context) (Stats, error) {
	var stats Stats
	err := r.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users)::int,
			(SELECT COUNT(*) FROM bookings)::int,
			(SELECT COUNT(*) FROM parking_lots)::int,
			COALESCE((SELECT SUM(total_price) FROM bookings WHERE status IN ('active', 'completed')), 0)::float8
	`).Scan(&stats.TotalUsers, &stats.TotalBookings, &stats.TotalParkingLots, &stats.Revenue)
	return stats, err
}
