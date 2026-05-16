package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/username/parking-service/internal/model"
	"github.com/username/parking-service/internal/repository"
	"github.com/username/parking-service/internal/validator"
)

type ReviewService struct {
	reviews  reviewStore
	parkings parkingGetter
}

type reviewStore interface {
	Create(context.Context, repository.CreateReviewParams) (model.Review, error)
	ListByParkingLot(context.Context, uuid.UUID, int, int) ([]model.Review, error)
}

type ReviewInput struct {
	Rating  int     `json:"rating"`
	Comment *string `json:"comment"`
}

func NewReviewService(reviews reviewStore, parkings parkingGetter) *ReviewService {
	return &ReviewService{reviews: reviews, parkings: parkings}
}

func (s *ReviewService) Create(ctx context.Context, actor Claims, lotID uuid.UUID, input ReviewInput) (model.Review, error) {
	if !validator.ValidateRating(input.Rating) {
		return model.Review{}, ValidationError("rating", "rating must be between 1 and 5")
	}
	if _, err := s.parkings.GetByID(ctx, lotID); errors.Is(err, repository.ErrNotFound) {
		return model.Review{}, ErrNotFound
	} else if err != nil {
		return model.Review{}, err
	}

	review, err := s.reviews.Create(ctx, repository.CreateReviewParams{
		UserID:       actor.UserID,
		ParkingLotID: lotID,
		Rating:       input.Rating,
		Comment:      input.Comment,
	})
	if errors.Is(err, repository.ErrConflict) {
		return model.Review{}, ErrConflict
	}
	return review, err
}

func (s *ReviewService) List(ctx context.Context, lotID uuid.UUID, limit int, offset int) ([]model.Review, error) {
	if _, err := s.parkings.GetByID(ctx, lotID); errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return s.reviews.ListByParkingLot(ctx, lotID, limit, offset)
}
