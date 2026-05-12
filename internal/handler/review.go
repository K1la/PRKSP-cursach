package handler

import (
	"net/http"

	"github.com/username/parking-service/internal/service"
)

func (api *API) createReview(w http.ResponseWriter, r *http.Request) {
	actor, _ := ActorFromContext(r.Context())
	lotID, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid parking lot id", "id")
		return
	}

	var input service.ReviewInput
	if err := decodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", "")
		return
	}

	review, err := api.reviews.Create(r.Context(), actor, lotID, input)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, review)
}

func (api *API) listReviews(w http.ResponseWriter, r *http.Request) {
	lotID, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid parking lot id", "id")
		return
	}

	limit, offset := pagination(r)
	reviews, err := api.reviews.List(r.Context(), lotID, limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, reviews)
}
