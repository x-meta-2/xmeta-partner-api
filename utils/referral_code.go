package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

const (
	referralCodeMinLen     = 5
	referralCodeMaxLen     = 7
	referralCodeDefaultLen = 7
	referralCodeCharset    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// GenerateReferralCode produces a 7-char uppercase alphanumeric code,
// e.g. "AB12CDE". No prefix — partners share the bare code.
func GenerateReferralCode() (string, error) {
	code := make([]byte, referralCodeDefaultLen)
	for i := 0; i < referralCodeDefaultLen; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(referralCodeCharset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate referral code: %w", err)
		}
		code[i] = referralCodeCharset[idx.Int64()]
	}
	return string(code), nil
}

// ValidateReferralCode checks a partner-supplied custom code: 5–7 chars,
// uppercase A-Z and 0-9 only.
func ValidateReferralCode(code string) error {
	if len(code) < referralCodeMinLen || len(code) > referralCodeMaxLen {
		return fmt.Errorf("referral code must be %d–%d characters", referralCodeMinLen, referralCodeMaxLen)
	}
	if code != strings.ToUpper(code) {
		return fmt.Errorf("referral code must be uppercase")
	}
	for _, c := range code {
		if !strings.ContainsRune(referralCodeCharset, c) {
			return fmt.Errorf("referral code must contain only A-Z and 0-9")
		}
	}
	return nil
}
