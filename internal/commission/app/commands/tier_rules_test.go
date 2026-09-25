package commands

import (
	"testing"

	"xmeta-partner/database"

	"github.com/stretchr/testify/require"
)

func TestBestTierForMetricsHonorsMaxVolume(t *testing.T) {
	maxVolume := 150.0
	tiers := []database.PartnerTier{
		{Base: database.Base{ID: "standard"}, Level: 1, MinVolume: 0},
		{Base: database.Base{ID: "silver"}, Level: 2, MinVolume: 100, MaxVolume: &maxVolume},
	}

	tier := bestTierForMetrics(tiers, 200, 0)

	require.Equal(t, "standard", tier.ID)
}

func TestBestTierForMetricsSelectsTierInsideVolumeRange(t *testing.T) {
	maxVolume := 150.0
	tiers := []database.PartnerTier{
		{Base: database.Base{ID: "standard"}, Level: 1, MinVolume: 0},
		{Base: database.Base{ID: "silver"}, Level: 2, MinVolume: 100, MaxVolume: &maxVolume},
	}

	tier := bestTierForMetrics(tiers, 125, 0)

	require.Equal(t, "silver", tier.ID)
}
