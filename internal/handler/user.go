package handler

import (
	"net/http"

	"github.com/username/parking-service/internal/service"
)

func (api *API) me(w http.ResponseWriter, r *http.Request) {
	actor, _ := ActorFromContext(r.Context())
	user, err := api.users.Get(r.Context(), actor.UserID)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, user)
}

func (api *API) updateMe(w http.ResponseWriter, r *http.Request) {
	actor, _ := ActorFromContext(r.Context())
	var input service.UpdateProfileInput
	if err := decodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", "")
		return
	}

	user, err := api.users.UpdateProfile(r.Context(), actor.UserID, input)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, user)
}

func (api *API) listUsers(w http.ResponseWriter, r *http.Request) {
	limit, offset := pagination(r)
	users, err := api.users.List(r.Context(), limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, users)
}

func (api *API) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid user id", "id")
		return
	}

	user, err := api.users.Get(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, user)
}

func (api *API) updateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid user id", "id")
		return
	}

	var input service.UpdateUserInput
	if err := decodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", "")
		return
	}

	user, err := api.users.Update(r.Context(), id, input)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, user)
}

func (api *API) deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid user id", "id")
		return
	}

	if err := api.users.Delete(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
