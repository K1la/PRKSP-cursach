package model

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleOwner UserRole = "owner"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Phone        *string   `json:"phone,omitempty"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ParkingLot struct {
	ID           uuid.UUID `json:"id"`
	OwnerID      uuid.UUID `json:"owner_id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description,omitempty"`
	Address      string    `json:"address"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	TotalSpots   int       `json:"total_spots"`
	PricePerHour float64   `json:"price_per_hour"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SpotType string

const (
	SpotStandard SpotType = "standard"
	SpotDisabled SpotType = "disabled"
	SpotElectric SpotType = "electric"
	SpotVIP      SpotType = "vip"
)

type ParkingSpot struct {
	ID           uuid.UUID `json:"id"`
	ParkingLotID uuid.UUID `json:"parking_lot_id"`
	SpotNumber   string    `json:"spot_number"`
	SpotType     SpotType  `json:"spot_type"`
	IsAvailable  bool      `json:"is_available"`
	Floor        *int      `json:"floor,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type BookingStatus string

const (
	BookingPending   BookingStatus = "pending"
	BookingActive    BookingStatus = "active"
	BookingCompleted BookingStatus = "completed"
	BookingCancelled BookingStatus = "cancelled"
)

type Booking struct {
	ID            uuid.UUID     `json:"id"`
	UserID        uuid.UUID     `json:"user_id"`
	ParkingSpotID uuid.UUID     `json:"parking_spot_id"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Status        BookingStatus `json:"status"`
	TotalPrice    float64       `json:"total_price"`
	VehiclePlate  string        `json:"vehicle_plate"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type Review struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	ParkingLotID uuid.UUID `json:"parking_lot_id"`
	Rating       int       `json:"rating"`
	Comment      *string   `json:"comment,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
