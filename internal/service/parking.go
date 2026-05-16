package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/username/parking-service/internal/model"
	"github.com/username/parking-service/internal/repository"
	"github.com/username/parking-service/internal/validator"
)

type ParkingService struct {
	parkings parkingStore
}

type parkingStore interface {
	List(context.Context, repository.ParkingLotFilter) ([]model.ParkingLot, error)
	GetByID(context.Context, uuid.UUID) (model.ParkingLot, error)
	Create(context.Context, repository.CreateParkingLotParams) (model.ParkingLot, error)
	Update(context.Context, uuid.UUID, repository.UpdateParkingLotParams) (model.ParkingLot, error)
	Delete(context.Context, uuid.UUID) error
	ListSpots(context.Context, uuid.UUID) ([]model.ParkingSpot, error)
}

type ParkingLotInput struct {
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	Address      string  `json:"address"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	TotalSpots   int     `json:"total_spots"`
	PricePerHour float64 `json:"price_per_hour"`
	IsActive     *bool   `json:"is_active"`
}

type ParkingFilter struct {
	Query     string
	Latitude  *float64
	Longitude *float64
	RadiusKM  *float64
	MaxPrice  *float64
	SpotType  *model.SpotType
	Limit     int
	Offset    int
}

func NewParkingService(parkings parkingStore) *ParkingService {
	return &ParkingService{parkings: parkings}
}

func (s *ParkingService) List(ctx context.Context, filter ParkingFilter) ([]model.ParkingLot, error) {
	hasGeo := filter.Latitude != nil || filter.Longitude != nil || filter.RadiusKM != nil
	if hasGeo {
		if filter.Latitude == nil || filter.Longitude == nil || filter.RadiusKM == nil {
			return nil, ValidationError("location", "lat, lng and radius must be provided together")
		}
		if !validator.ValidateCoordinates(*filter.Latitude, *filter.Longitude) {
			return nil, ValidationError("coordinates", "invalid coordinates")
		}
		if *filter.RadiusKM <= 0 {
			return nil, ValidationError("radius", "radius must be positive")
		}
	}
	if filter.MaxPrice != nil && *filter.MaxPrice < 0 {
		return nil, ValidationError("max_price", "max_price must be non-negative")
	}
	if filter.SpotType != nil && *filter.SpotType != model.SpotStandard && *filter.SpotType != model.SpotDisabled && *filter.SpotType != model.SpotElectric && *filter.SpotType != model.SpotVIP {
		return nil, ValidationError("spot_type", "invalid spot type")
	}

	return s.parkings.List(ctx, repository.ParkingLotFilter{
		Query:     filter.Query,
		Latitude:  filter.Latitude,
		Longitude: filter.Longitude,
		RadiusKM:  filter.RadiusKM,
		MaxPrice:  filter.MaxPrice,
		SpotType:  filter.SpotType,
		Limit:     filter.Limit,
		Offset:    filter.Offset,
	})
}

func (s *ParkingService) Get(ctx context.Context, id uuid.UUID) (model.ParkingLot, error) {
	lot, err := s.parkings.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return model.ParkingLot{}, ErrNotFound
	}
	return lot, err
}

func (s *ParkingService) Create(ctx context.Context, actor Claims, input ParkingLotInput) (model.ParkingLot, error) {
	if actor.Role != model.RoleOwner && actor.Role != model.RoleAdmin {
		return model.ParkingLot{}, ErrForbidden
	}
	if err := validateParkingLotInput(input); err != nil {
		return model.ParkingLot{}, err
	}
	return s.parkings.Create(ctx, repository.CreateParkingLotParams{
		OwnerID:      actor.UserID,
		Name:         strings.TrimSpace(input.Name),
		Description:  input.Description,
		Address:      strings.TrimSpace(input.Address),
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		TotalSpots:   input.TotalSpots,
		PricePerHour: input.PricePerHour,
	})
}

func (s *ParkingService) Update(ctx context.Context, actor Claims, id uuid.UUID, input ParkingLotInput) (model.ParkingLot, error) {
	lot, err := s.Get(ctx, id)
	if err != nil {
		return model.ParkingLot{}, err
	}
	if !canManageParking(actor, lot) {
		return model.ParkingLot{}, ErrForbidden
	}
	if err := validateParkingLotInput(input); err != nil {
		return model.ParkingLot{}, err
	}

	isActive := lot.IsActive
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	updated, err := s.parkings.Update(ctx, id, repository.UpdateParkingLotParams{
		Name:         strings.TrimSpace(input.Name),
		Description:  input.Description,
		Address:      strings.TrimSpace(input.Address),
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		TotalSpots:   input.TotalSpots,
		PricePerHour: input.PricePerHour,
		IsActive:     isActive,
	})
	if errors.Is(err, repository.ErrNotFound) {
		return model.ParkingLot{}, ErrNotFound
	}
	return updated, err
}

func (s *ParkingService) Delete(ctx context.Context, actor Claims, id uuid.UUID) error {
	lot, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if !canManageParking(actor, lot) {
		return ErrForbidden
	}
	err = s.parkings.Delete(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *ParkingService) ListSpots(ctx context.Context, id uuid.UUID) ([]model.ParkingSpot, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}
	return s.parkings.ListSpots(ctx, id)
}

func canManageParking(actor Claims, lot model.ParkingLot) bool {
	return actor.Role == model.RoleAdmin || (actor.Role == model.RoleOwner && actor.UserID == lot.OwnerID)
}

func validateParkingLotInput(input ParkingLotInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return ValidationError("name", "name is required")
	}
	if strings.TrimSpace(input.Address) == "" {
		return ValidationError("address", "address is required")
	}
	if !validator.ValidateCoordinates(input.Latitude, input.Longitude) {
		return ValidationError("coordinates", "invalid coordinates")
	}
	if input.TotalSpots <= 0 {
		return ValidationError("total_spots", "total_spots must be positive")
	}
	if input.PricePerHour < 0 {
		return ValidationError("price_per_hour", "price_per_hour must be non-negative")
	}
	return nil
}
