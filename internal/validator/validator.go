package validator

import (
	"net/mail"
	"regexp"
	"strings"
	"time"
)

var (
	phonePattern        = regexp.MustCompile(`^\+?[0-9\s\-()]{7,20}$`)
	vehiclePlatePattern = regexp.MustCompile(`^[A-Za-zА-Яа-я0-9\- ]{3,20}$`)
)

func ValidateEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" || len(email) > 255 {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

func ValidatePassword(password string) bool {
	return len(password) >= 8
}

func ValidatePhone(phone string) bool {
	phone = strings.TrimSpace(phone)
	return phone == "" || phonePattern.MatchString(phone)
}

func ValidateVehiclePlate(plate string) bool {
	return vehiclePlatePattern.MatchString(strings.TrimSpace(plate))
}

func ValidateCoordinates(latitude float64, longitude float64) bool {
	return latitude >= -90 && latitude <= 90 && longitude >= -180 && longitude <= 180
}

func ValidateBookingTime(start time.Time, end time.Time, now time.Time) bool {
	return start.After(now) && end.After(start)
}

func ValidateRating(rating int) bool {
	return rating >= 1 && rating <= 5
}
