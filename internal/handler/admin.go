package handler

import "net/http"

func (api *API) adminStats(w http.ResponseWriter, r *http.Request) {
	actor, _ := ActorFromContext(r.Context())
	stats, err := api.admin.Stats(r.Context(), actor)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}
