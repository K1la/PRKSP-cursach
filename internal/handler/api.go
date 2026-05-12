package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/username/parking-service/internal/model"
	"github.com/username/parking-service/internal/service"
)

type API struct {
	auth     *service.AuthService
	users    *service.UserService
	parkings *service.ParkingService
	bookings *service.BookingService
	reviews  *service.ReviewService
	admin    *service.AdminService
}

type contextKey string

const actorContextKey contextKey = "actor"

func NewAPI(
	auth *service.AuthService,
	users *service.UserService,
	parkings *service.ParkingService,
	bookings *service.BookingService,
	reviews *service.ReviewService,
	admin *service.AdminService,
) *API {
	return &API{
		auth:     auth,
		users:    users,
		parkings: parkings,
		bookings: bookings,
		reviews:  reviews,
		admin:    admin,
	}
}

func (api *API) Routes(r chi.Router) {
	r.Post("/auth/register", api.register)
	r.Post("/auth/login", api.login)
	r.Post("/auth/refresh", api.refresh)

	r.Get("/parking-lots", api.listParkingLots)
	r.Get("/parking-lots/{id}", api.getParkingLot)
	r.Get("/parking-lots/{id}/spots", api.listParkingSpots)
	r.Get("/parking-lots/{id}/reviews", api.listReviews)

	r.Group(func(r chi.Router) {
		r.Use(api.RequireAuth)

		r.Get("/users/me", api.me)
		r.Put("/users/me", api.updateMe)

		r.Get("/bookings", api.listMyBookings)
		r.Post("/bookings", api.createBooking)
		r.Get("/bookings/{id}", api.getBooking)
		r.Put("/bookings/{id}/cancel", api.cancelBooking)

		r.Post("/parking-lots", api.createParkingLot)
		r.Put("/parking-lots/{id}", api.updateParkingLot)
		r.Delete("/parking-lots/{id}", api.deleteParkingLot)
		r.Get("/parking-lots/{id}/bookings", api.listParkingBookings)
		r.Post("/parking-lots/{id}/reviews", api.createReview)

		r.Group(func(r chi.Router) {
			r.Use(api.RequireRole(model.RoleAdmin))
			r.Get("/users", api.listUsers)
			r.Get("/users/{id}", api.getUser)
			r.Put("/users/{id}", api.updateUser)
			r.Delete("/users/{id}", api.deleteUser)
			r.Get("/admin/stats", api.adminStats)
		})
	})
}

func (api *API) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" || header == token {
			WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing bearer token", "")
			return
		}

		claims, err := api.auth.ParseToken(token, "access")
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token", "")
			return
		}

		ctx := context.WithValue(r.Context(), actorContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (api *API) RequireRole(roles ...model.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, ok := ActorFromContext(r.Context())
			if !ok {
				WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing auth context", "")
				return
			}
			for _, role := range roles {
				if actor.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			WriteError(w, http.StatusForbidden, "FORBIDDEN", "insufficient role", "")
		})
	}
}

func ActorFromContext(ctx context.Context) (service.Claims, bool) {
	actor, ok := ctx.Value(actorContextKey).(service.Claims)
	return actor, ok
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, name))
}

func pagination(r *http.Request) (int, int) {
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 20)
	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	if limit > 100 {
		limit = 100
	}
	return limit, (page - 1) * limit
}

func parsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func parseOptionalFloat(raw string) (*float64, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func handleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var fieldErr service.FieldError
	if errors.As(err, &fieldErr) {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", fieldErr.Message, fieldErr.Field)
		return
	}

	switch {
	case errors.Is(err, service.ErrUnauthorized):
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", "")
	case errors.Is(err, service.ErrForbidden):
		WriteError(w, http.StatusForbidden, "FORBIDDEN", "access denied", "")
	case errors.Is(err, service.ErrNotFound):
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "resource not found", "")
	case errors.Is(err, service.ErrConflict):
		WriteError(w, http.StatusConflict, "CONFLICT", "operation conflicts with current state", "")
	default:
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
	}
}
