package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/username/parking-service/internal/model"
	"github.com/username/parking-service/internal/repository"
	"github.com/username/parking-service/internal/validator"
)

type BookingService struct {
	bookings *repository.BookingRepository
	parkings *repository.ParkingRepository
}

type BookingInput struct {
	ParkingSpotID uuid.UUID `json:"parking_spot_id"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	VehiclePlate  string    `json:"vehicle_plate"`
}

func NewBookingService(bookings *repository.BookingRepository, parkings *repository.ParkingRepository) *BookingService {
	return &BookingService{bookings: bookings, parkings: parkings}
}

func (s *BookingService) Create(ctx context.Context, actor Claims, input BookingInput) (model.Booking, error) {
	if input.ParkingSpotID == uuid.Nil {
		return model.Booking{}, ValidationError("parking_spot_id", "parking_spot_id is required")
	}
	if !validator.ValidateBookingTime(input.StartTime, input.EndTime, time.Now()) {
		return model.Booking{}, ValidationError("time", "end_time must be after start_time and start_time must be in future")
	}
	input.VehiclePlate = strings.TrimSpace(input.VehiclePlate)
	if !validator.ValidateVehiclePlate(input.VehiclePlate) {
		return model.Booking{}, ValidationError("vehicle_plate", "invalid vehicle plate")
	}

	info, err := s.bookings.GetSpotBookingInfo(ctx, input.ParkingSpotID)
	if errors.Is(err, repository.ErrNotFound) {
		return model.Booking{}, ErrNotFound
	}
	if err != nil {
		return model.Booking{}, err
	}
	if !info.LotActive || !info.IsAvailable {
		return model.Booking{}, ErrConflict
	}

	overlap, err := s.bookings.HasOverlap(ctx, input.ParkingSpotID, input.StartTime, input.EndTime)
	if err != nil {
		return model.Booking{}, err
	}
	if overlap {
		return model.Booking{}, ErrConflict
	}

	hours := math.Ceil(input.EndTime.Sub(input.StartTime).Hours()*100) / 100
	totalPrice := hours * info.PricePerHour

	return s.bookings.Create(ctx, repository.CreateBookingParams{
		UserID:        actor.UserID,
		ParkingSpotID: input.ParkingSpotID,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		TotalPrice:    totalPrice,
		VehiclePlate:  input.VehiclePlate,
	})
}

func (s *BookingService) ListMine(ctx context.Context, actor Claims, limit int, offset int) ([]model.Booking, error) {
	return s.bookings.ListByUser(ctx, actor.UserID, limit, offset)
}

func (s *BookingService) Get(ctx context.Context, actor Claims, id uuid.UUID) (model.Booking, error) {
	booking, err := s.bookings.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return model.Booking{}, ErrNotFound
	}
	if err != nil {
		return model.Booking{}, err
	}
	if actor.Role != model.RoleAdmin && booking.UserID != actor.UserID {
		return model.Booking{}, ErrForbidden
	}
	return booking, nil
}

func (s *BookingService) Cancel(ctx context.Context, actor Claims, id uuid.UUID) (model.Booking, error) {
	booking, err := s.Get(ctx, actor, id)
	if err != nil {
		return model.Booking{}, err
	}
	if actor.Role != model.RoleAdmin && booking.UserID != actor.UserID {
		return model.Booking{}, ErrForbidden
	}

	booking, err = s.bookings.Cancel(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return model.Booking{}, ErrConflict
	}
	return booking, err
}

func (s *BookingService) ListForParkingLot(ctx context.Context, actor Claims, lotID uuid.UUID, limit int, offset int) ([]model.Booking, error) {
	lot, err := s.parkings.GetByID(ctx, lotID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if actor.Role != model.RoleAdmin && !(actor.Role == model.RoleOwner && actor.UserID == lot.OwnerID) {
		return nil, ErrForbidden
	}
	return s.bookings.ListByParkingLot(ctx, lotID, limit, offset)
}
