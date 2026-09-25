package commands

import (
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/commission/port"
)

func ulaanbaatarLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Ulaanbaatar")
	if err != nil {
		return time.Local
	}
	return location
}

func monthRange(t time.Time) (time.Time, time.Time) {
	location := ulaanbaatarLocation()
	local := t.In(location)
	startedAt := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, location)
	return startedAt, startedAt.AddDate(0, 1, 0)
}

func previousMonthRange(t time.Time) (time.Time, time.Time) {
	currentMonth, _ := monthRange(t)
	startedAt := currentMonth.AddDate(0, -1, 0)
	return startedAt, currentMonth
}

func sameMonth(a, b time.Time) bool {
	aStart, _ := monthRange(a)
	bStart, _ := monthRange(b)
	return aStart.Equal(bStart)
}

func bestTierForMetrics(tiers []database.PartnerTier, totalVolume float64, activeClients int64) *database.PartnerTier {
	var bestTier *database.PartnerTier
	for i := range tiers {
		tier := &tiers[i]
		volumeOK := tier.MinVolume == 0 || totalVolume >= tier.MinVolume
		maxVolumeOK := tier.MaxVolume == nil || totalVolume < *tier.MaxVolume
		clientsOK := tier.MinActiveClients == 0 || activeClients >= int64(tier.MinActiveClients)
		if volumeOK && maxVolumeOK && clientsOK {
			bestTier = tier
		}
	}
	return bestTier
}

func maybeUpgradeTierForTrade(txRepo port.TradeEventRepo, partner *database.Partner, tradeDate time.Time) error {
	if !sameMonth(tradeDate, time.Now()) {
		return nil
	}

	startedAt, endedAt := monthRange(tradeDate)
	totalVolume, err := txRepo.GetPartnerTotalVolume(partner.ID, startedAt, endedAt)
	if err != nil {
		return err
	}

	activeClients, err := txRepo.GetPartnerActiveClients(partner.ID, startedAt, endedAt)
	if err != nil {
		return err
	}

	tiers, err := txRepo.FindAllTiersAsc()
	if err != nil {
		return err
	}

	bestTier := bestTierForMetrics(tiers, totalVolume, activeClients)
	if bestTier != nil && bestTier.Level > partner.Tier.Level {
		return txRepo.UpgradePartnerTier(partner.ID, bestTier.ID, bestTier.Level)
	}
	return nil
}
