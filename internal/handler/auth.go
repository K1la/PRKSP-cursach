package handler

import (
	"net/http"

	"github.com/username/parking-service/internal/service"
)

func (api *API) register(w http.ResponseWriter, r *http.Request) {
	var input service.RegisterInput
	if err := decodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", "")
		return
	}

	response, err := api.auth.Register(r.Context(), input)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, response)
}

func (api *API) login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := decodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", "")
		return
	}

	response, err := api.auth.Login(r.Context(), input)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, response)
}

func (api *API) refresh(w http.ResponseWriter, r *http.Request) {
	var input service.RefreshInput
	if err := decodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", "")
		return
	}

	tokens, err := api.auth.Refresh(r.Context(), input)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, tokens)
}
