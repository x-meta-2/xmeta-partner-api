package utils

import "strings"

// MaskEmail produces "it***@gmail.com" from "itsoftware027@gmail.com".
// Keeps the first 2 chars of the local part + the full domain so callers
// can tell records apart without seeing the whole address.
func MaskEmail(email string) string {
	at := strings.LastIndex(email, "@")
	if at <= 0 {
		return email
	}
	local, domain := email[:at], email[at:]
	if len(local) <= 2 {
		return local + "***" + domain
	}
	return local[:2] + "***" + domain
}

// LastInitial returns "B." for "Battugs". Empty string for empty input.
func LastInitial(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	r := []rune(name)
	return string(r[0]) + "."
}
