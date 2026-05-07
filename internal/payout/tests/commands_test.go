package tests

import (
	"testing"

	"xmeta-partner/database"
	"xmeta-partner/internal/payout/app/queries"
	"xmeta-partner/internal/payout/domain"
	"xmeta-partner/structs"

	"github.com/stretchr/testify/assert"
)

// ── Domain constants ──

func TestPayoutStatus_Constants(t *testing.T) {
	assert.Equal(t, domain.PayoutStatus("pending"), domain.StatusPending)
	assert.Equal(t, domain.PayoutStatus("processing"), domain.StatusProcessing)
	assert.Equal(t, domain.PayoutStatus("completed"), domain.StatusCompleted)
	assert.Equal(t, domain.PayoutStatus("failed"), domain.StatusFailed)
}

func TestPayoutStatus_MatchesDatabase(t *testing.T) {
	assert.Equal(t, string(domain.StatusPending), string(database.PayoutStatusPending))
	assert.Equal(t, string(domain.StatusProcessing), string(database.PayoutStatusProcessing))
	assert.Equal(t, string(domain.StatusCompleted), string(database.PayoutStatusCompleted))
	assert.Equal(t, string(domain.StatusFailed), string(database.PayoutStatusFailed))
}

func TestErrPayoutNotFound(t *testing.T) {
	assert.EqualError(t, domain.ErrPayoutNotFound, "payout not found or already processed")
}

func TestMinPayoutAmount(t *testing.T) {
	assert.Equal(t, 10.0, queries.MinPayoutAmount)
}

func TestPayoutReviewParams(t *testing.T) {
	params := structs.PayoutReviewParams{FailureReason: "insufficient documents"}
	assert.Equal(t, "insufficient documents", params.FailureReason)
}
