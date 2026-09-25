package commands

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSyncRange_DateOnlyUsesUlaanbaatarDay(t *testing.T) {
	startedAt, endedAt, err := parseSyncRange("2026-09-24", "2026-09-24")

	require.NoError(t, err)
	location, err := time.LoadLocation("Asia/Ulaanbaatar")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 9, 24, 0, 0, 0, 0, location), startedAt)
	assert.Equal(t, time.Date(2026, 9, 25, 0, 0, 0, 0, location), endedAt)
}

func TestParseSyncRange_RFC3339KeepsExplicitInstant(t *testing.T) {
	startedAt, endedAt, err := parseSyncRange("2026-09-24T00:00:00Z", "2026-09-25T00:00:00Z")

	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC), startedAt)
	assert.Equal(t, time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC), endedAt)
}

func TestTruncate4AlwaysRoundsDown(t *testing.T) {
	assert.Equal(t, 12.3456, truncate4(12.34569))
	assert.Equal(t, 0.0001, truncate4(0.00019999))
	assert.Equal(t, 0.0, truncate4(0.00009999))
}
