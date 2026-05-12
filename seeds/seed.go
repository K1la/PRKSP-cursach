package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type seedUser struct {
	Email    string
	Password string
	Name     string
	Role     string
}

type seedParkingLot struct {
	OwnerEmail   string
	Name         string
	Description  string
	Address      string
	Latitude     float64
	Longitude    float64
	TotalSpots   int
	PricePerHour float64
}

var users = []seedUser{
	{"admin@parkease.ru", "password123", "Администратор Сервиса", "admin"},
	{"owner1@parkease.ru", "password123", "Иван Парковкин", "owner"},
	{"owner2@parkease.ru", "password123", "Мария Гаражева", "owner"},
	{"user1@parkease.ru", "password123", "Алексей Водителев", "user"},
	{"user2@parkease.ru", "password123", "Елена Автомобилева", "user"},
	{"user3@parkease.ru", "password123", "Дмитрий Проездов", "user"},
}

var parkingLots = []seedParkingLot{
	{"owner1@parkease.ru", "ParkEase у Кремля", "Парковка рядом с центром города", "Манежная площадь, Москва", 55.7520, 37.6175, 50, 200},
	{"owner1@parkease.ru", "Парковка ТЦ Европейский", "Крытая парковка у торгового центра", "Площадь Киевского Вокзала, 2", 55.7449, 37.5652, 120, 150},
	{"owner1@parkease.ru", "Гараж на Тверской", "Компактный городской гараж", "Тверская улица, Москва", 55.7650, 37.6058, 30, 300},
	{"owner2@parkease.ru", "Парковка ВДНХ", "Большая открытая парковка", "Проспект Мира, 119", 55.8263, 37.6376, 200, 80},
	{"owner2@parkease.ru", "Подземная у Арбата", "Подземная охраняемая парковка", "Новый Арбат, Москва", 55.7522, 37.5876, 45, 250},
	{"owner2@parkease.ru", "Парковка Сити", "Парковка рядом с деловым центром", "Пресненская набережная, Москва", 55.7494, 37.5374, 500, 180},
	{"owner1@parkease.ru", "Гараж Сокольники", "Парковка у парка", "Сокольническая площадь, Москва", 55.7898, 37.6797, 60, 100},
	{"owner2@parkease.ru", "Открытая Лужники", "Большая парковка у стадиона", "Лужнецкая набережная, Москва", 55.7155, 37.5539, 300, 120},
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/parking?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		logger.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		logger.Error("database is not reachable", "error", err)
		os.Exit(1)
	}

	if err := seed(ctx, db); err != nil {
		logger.Error("seed failed", "error", err)
		os.Exit(1)
	}

	logger.Info("seed completed")
}

func seed(ctx context.Context, db *pgxpool.Pool) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, user := range users {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO users (email, password_hash, name, role)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (email) DO UPDATE
			SET password_hash = EXCLUDED.password_hash,
			    name = EXCLUDED.name,
			    role = EXCLUDED.role,
			    updated_at = NOW()
		`, user.Email, string(passwordHash), user.Name, user.Role)
		if err != nil {
			return err
		}
	}

	for _, lot := range parkingLots {
		var lotID string
		err := tx.QueryRow(ctx, `SELECT id FROM parking_lots WHERE name = $1`, lot.Name).Scan(&lotID)
		if err != nil {
			err = tx.QueryRow(ctx, `
			INSERT INTO parking_lots (
				owner_id, name, description, address, latitude, longitude, total_spots, price_per_hour
			)
			SELECT id, $2, $3, $4, $5, $6, $7, $8
			FROM users
			WHERE email = $1
			RETURNING id
		`, lot.OwnerEmail, lot.Name, lot.Description, lot.Address, lot.Latitude, lot.Longitude, lot.TotalSpots, lot.PricePerHour).Scan(&lotID)
			if err != nil {
				return err
			}
		}

		for spot := 1; spot <= min(lot.TotalSpots, 20); spot++ {
			spotType := "standard"
			if spot%15 == 0 {
				spotType = "electric"
			} else if spot%10 == 0 {
				spotType = "disabled"
			} else if spot%7 == 0 {
				spotType = "vip"
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO parking_spots (parking_lot_id, spot_number, spot_type, floor)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (parking_lot_id, spot_number) DO NOTHING
			`, lotID, formatSpotNumber(spot), spotType, 1)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func formatSpotNumber(number int) string {
	return fmt.Sprintf("A%02d", number)
}
