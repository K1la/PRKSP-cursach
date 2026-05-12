package handler

import (
	"net/http"

	"github.com/username/parking-service/internal/service"
)

func (api *API) createBooking(w http.ResponseWriter, r *http.Request) {
	actor, _ := ActorFromContext(r.Context())
	var input service.BookingInput
	if err := decodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", "")
		return
	}

	booking, err := api.bookings.Create(r.Context(), actor, input)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, booking)
}

func (api *API) listMyBookings(w http.ResponseWriter, r *http.Request) {
	actor, _ := ActorFromContext(r.Context())
	limit, offset := pagination(r)
	bookings, err := api.bookings.ListMine(r.Context(), actor, limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, bookings)
}

func (api *API) getBooking(w http.ResponseWriter, r *http.Request) {
	actor, _ := ActorFromContext(r.Context())
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid booking id", "id")
		return
	}

	booking, err := api.bookings.Get(r.Context(), actor, id)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

func (api *API) cancelBooking(w http.ResponseWriter, r *http.Request) {
	actor, _ := ActorFromContext(r.Context())
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid booking id", "id")
		return
	}

	booking, err := api.bookings.Cancel(r.Context(), actor, id)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, booking)
}

func (api *API) listParkingBookings(w http.ResponseWriter, r *http.Request) {
	actor, _ := ActorFromContext(r.Context())
	lotID, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid parking lot id", "id")
		return
	}

	limit, offset := pagination(r)
	bookings, err := api.bookings.ListForParkingLot(r.Context(), actor, lotID, limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, bookings)
}
