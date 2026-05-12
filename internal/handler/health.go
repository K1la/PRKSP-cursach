package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) HealthHandler {
	return HealthHandler{db: db}
}

func (h HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "connected"
	status := http.StatusOK
	if err := h.db.Ping(ctx); err != nil {
		dbStatus = "unavailable"
		status = http.StatusServiceUnavailable
	}

	WriteJSON(w, status, map[string]string{
		"status":  "ok",
		"db":      dbStatus,
		"version": "1.0.0",
	})
}
