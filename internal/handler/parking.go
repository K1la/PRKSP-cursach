package handler

import (
	"net/http"

	"github.com/username/parking-service/internal/service"
)

func (api *API) listParkingLots(w http.ResponseWriter, r *http.Request) {
	limit, offset := pagination(r)
	lat, err := parseOptionalFloat(r.URL.Query().Get("lat"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid latitude", "lat")
		return
	}
	lng, err := parseOptionalFloat(r.URL.Query().Get("lng"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid longitude", "lng")
		return
	}
	radius, err := parseOptionalFloat(r.URL.Query().Get("radius"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid radius", "radius")
		return
	}

	lots, err := api.parkings.List(r.Context(), service.ParkingFilter{
		Query:     r.URL.Query().Get("q"),
		Latitude:  lat,
		Longitude: lng,
		RadiusKM:  radius,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, lots)
}

func (api *API) getParkingLot(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid parking lot id", "id")
		return
	}

	lot, err := api.parkings.Get(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, lot)
}

func (api *API) createParkingLot(w http.ResponseWriter, r *http.Request) {
	actor, _ := ActorFromContext(r.Context())
	var input service.ParkingLotInput
	if err := decodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", "")
		return
	}

	lot, err := api.parkings.Create(r.Context(), actor, input)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, lot)
}

func (api *API) updateParkingLot(w http.ResponseWriter, r *http.Request) {
	actor, _ := ActorFromContext(r.Context())
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid parking lot id", "id")
		return
	}

	var input service.ParkingLotInput
	if err := decodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", "")
		return
	}

	lot, err := api.parkings.Update(r.Context(), actor, id, input)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, lot)
}

func (api *API) deleteParkingLot(w http.ResponseWriter, r *http.Request) {
	actor, _ := ActorFromContext(r.Context())
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid parking lot id", "id")
		return
	}

	if err := api.parkings.Delete(r.Context(), actor, id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) listParkingSpots(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid parking lot id", "id")
		return
	}

	spots, err := api.parkings.ListSpots(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, spots)
}
