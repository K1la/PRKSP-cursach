package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/username/parking-service/internal/model"
	"github.com/username/parking-service/internal/repository"
)

type fakeParkingStore struct {
	filter repository.ParkingLotFilter
	lot    model.ParkingLot
}

func (s *fakeParkingStore) List(_ context.Context, filter repository.ParkingLotFilter) ([]model.ParkingLot, error) {
	s.filter = filter
	return []model.ParkingLot{s.lot}, nil
}
func (s *fakeParkingStore) GetByID(context.Context, uuid.UUID) (model.ParkingLot, error) {
	return s.lot, nil
}
func (s *fakeParkingStore) Create(context.Context, repository.CreateParkingLotParams) (model.ParkingLot, error) {
	return s.lot, nil
}
func (s *fakeParkingStore) Update(context.Context, uuid.UUID, repository.UpdateParkingLotParams) (model.ParkingLot, error) {
	return s.lot, nil
}
func (s *fakeParkingStore) Delete(context.Context, uuid.UUID) error { return nil }
func (s *fakeParkingStore) ListSpots(context.Context, uuid.UUID) ([]model.ParkingSpot, error) {
	return nil, nil
}

func TestParkingCreateRejectsUserRole(t *testing.T) {
	svc := NewParkingService(&fakeParkingStore{})
	_, err := svc.Create(context.Background(), Claims{UserID: uuid.New(), Role: model.RoleUser}, ParkingLotInput{})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Create() error = %v, want ErrForbidden", err)
	}
}

func TestParkingListPassesOptionalFilters(t *testing.T) {
	store := &fakeParkingStore{}
	svc := NewParkingService(store)
	maxPrice := 200.0
	spotType := model.SpotElectric
	_, err := svc.List(context.Background(), ParkingFilter{MaxPrice: &maxPrice, SpotType: &spotType, Limit: 20})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if store.filter.MaxPrice == nil || *store.filter.MaxPrice != maxPrice {
		t.Fatalf("MaxPrice filter not passed")
	}
	if store.filter.SpotType == nil || *store.filter.SpotType != spotType {
		t.Fatalf("SpotType filter not passed")
	}
}
