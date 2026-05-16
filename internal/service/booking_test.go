package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/username/parking-service/internal/model"
	"github.com/username/parking-service/internal/repository"
)

type fakeBookingStore struct {
	info    repository.SpotBookingInfo
	overlap bool
	created repository.CreateBookingParams
	booking model.Booking
}

func (s *fakeBookingStore) GetSpotBookingInfo(context.Context, uuid.UUID) (repository.SpotBookingInfo, error) {
	return s.info, nil
}

func (s *fakeBookingStore) HasOverlap(context.Context, uuid.UUID, time.Time, time.Time) (bool, error) {
	return s.overlap, nil
}

func (s *fakeBookingStore) Create(_ context.Context, params repository.CreateBookingParams) (model.Booking, error) {
	s.created = params
	booking := s.booking
	booking.UserID = params.UserID
	booking.ParkingSpotID = params.ParkingSpotID
	booking.StartTime = params.StartTime
	booking.EndTime = params.EndTime
	booking.TotalPrice = params.TotalPrice
	booking.VehiclePlate = params.VehiclePlate
	return booking, nil
}

func (s *fakeBookingStore) GetByID(context.Context, uuid.UUID) (model.Booking, error) {
	return s.booking, nil
}

func (s *fakeBookingStore) ListByUser(context.Context, uuid.UUID, int, int) ([]model.Booking, error) {
	return []model.Booking{s.booking}, nil
}

func (s *fakeBookingStore) ListByParkingLot(context.Context, uuid.UUID, int, int) ([]model.Booking, error) {
	return []model.Booking{s.booking}, nil
}

func (s *fakeBookingStore) Cancel(context.Context, uuid.UUID) (model.Booking, error) {
	if s.booking.Status != model.BookingPending && s.booking.Status != model.BookingActive {
		return model.Booking{}, repository.ErrNotFound
	}
	s.booking.Status = model.BookingCancelled
	return s.booking, nil
}

type fakeParkingGetter struct {
	lot model.ParkingLot
	err error
}

func (s fakeParkingGetter) GetByID(context.Context, uuid.UUID) (model.ParkingLot, error) {
	return s.lot, s.err
}

func TestBookingCreateCalculatesPrice(t *testing.T) {
	userID := uuid.New()
	spotID := uuid.New()
	start := time.Now().Add(time.Hour)
	end := start.Add(90 * time.Minute)
	store := &fakeBookingStore{
		info:    repository.SpotBookingInfo{SpotID: spotID, IsAvailable: true, LotActive: true, PricePerHour: 200},
		booking: model.Booking{ID: uuid.New(), Status: model.BookingPending},
	}
	svc := NewBookingService(store, fakeParkingGetter{})

	booking, err := svc.Create(context.Background(), Claims{UserID: userID, Role: model.RoleUser}, BookingInput{
		ParkingSpotID: spotID,
		StartTime:     start,
		EndTime:       end,
		VehiclePlate:  "A123BC777",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if booking.TotalPrice != 300 {
		t.Fatalf("TotalPrice = %v, want 300", booking.TotalPrice)
	}
}

func TestBookingCreateRejectsOverlap(t *testing.T) {
	spotID := uuid.New()
	start := time.Now().Add(time.Hour)
	store := &fakeBookingStore{
		info:    repository.SpotBookingInfo{SpotID: spotID, IsAvailable: true, LotActive: true, PricePerHour: 100},
		overlap: true,
	}
	svc := NewBookingService(store, fakeParkingGetter{})

	_, err := svc.Create(context.Background(), Claims{UserID: uuid.New(), Role: model.RoleUser}, BookingInput{
		ParkingSpotID: spotID,
		StartTime:     start,
		EndTime:       start.Add(time.Hour),
		VehiclePlate:  "A123BC777",
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("Create() error = %v, want ErrConflict", err)
	}
}

func TestBookingListForParkingLotRejectsForeignOwner(t *testing.T) {
	lotID := uuid.New()
	svc := NewBookingService(&fakeBookingStore{}, fakeParkingGetter{lot: model.ParkingLot{ID: lotID, OwnerID: uuid.New()}})

	_, err := svc.ListForParkingLot(context.Background(), Claims{UserID: uuid.New(), Role: model.RoleOwner}, lotID, 20, 0)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("ListForParkingLot() error = %v, want ErrForbidden", err)
	}
}
