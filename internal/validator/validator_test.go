package validator

import (
	"testing"
	"time"
)

func TestValidateBookingTime(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		want  bool
	}{
		{"valid future range", now.Add(time.Hour), now.Add(2 * time.Hour), true},
		{"start in past", now.Add(-time.Hour), now.Add(time.Hour), false},
		{"end before start", now.Add(2 * time.Hour), now.Add(time.Hour), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateBookingTime(tt.start, tt.end, now); got != tt.want {
				t.Fatalf("ValidateBookingTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func FuzzValidateEmail(f *testing.F) {
	f.Add("test@example.com")
	f.Add("")
	f.Add("not-an-email")
	f.Fuzz(func(t *testing.T, email string) {
		_ = ValidateEmail(email)
	})
}

func FuzzValidatePhone(f *testing.F) {
	f.Add("+7 (999) 123-45-67")
	f.Add("")
	f.Add("abc")
	f.Fuzz(func(t *testing.T, phone string) {
		_ = ValidatePhone(phone)
	})
}

func FuzzValidateVehiclePlate(f *testing.F) {
	f.Add("A123BC777")
	f.Add("")
	f.Add("too-long-plate-value-123456")
	f.Fuzz(func(t *testing.T, plate string) {
		_ = ValidateVehiclePlate(plate)
	})
}

func FuzzParseBookingDate(f *testing.F) {
	f.Add("2026-05-15T12:00:00Z")
	f.Add("")
	f.Add("not-a-date")
	f.Fuzz(func(t *testing.T, raw string) {
		_, _ = time.Parse(time.RFC3339, raw)
	})
}
