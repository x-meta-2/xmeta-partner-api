package utils

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

const defaultReferralBaseURL = "https://x-meta.com/"

func BuildReferralURL(code string) string {
	base := strings.TrimSpace(viper.GetString("REFERRAL_BASE_URL"))
	if base == "" {
		base = defaultReferralBaseURL
	}

	u, err := url.Parse(base)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Sprintf("%s?ref=%s", strings.TrimRight(defaultReferralBaseURL, "/"), url.QueryEscape(code))
	}

	q := u.Query()
	q.Set("ref", code)
	u.RawQuery = q.Encode()
	return u.String()
}
