package commands

import (
	"time"

	"xmeta-partner/database"

	"gorm.io/gorm"
)

type MonthlyTierReviewHandler struct {
	DB *gorm.DB
}

type MonthlyTierReviewResult struct {
	Month        string                  `json:"month"`
	Total        int                     `json:"total"`
	Updated      int                     `json:"updated"`
	Unchanged    int                     `json:"unchanged"`
	Failed       int                     `json:"failed"`
	StartedAt    time.Time               `json:"startedAt"`
	EndedAt      time.Time               `json:"endedAt"`
	PartnerStats []MonthlyTierReviewItem `json:"partnerStats,omitempty"`
}

type MonthlyTierReviewItem struct {
	PartnerID     string  `json:"partnerId"`
	PreviousTier  string  `json:"previousTier"`
	AssignedTier  string  `json:"assignedTier"`
	TotalVolume   float64 `json:"totalVolume"`
	ActiveClients int64   `json:"activeClients"`
	Error         string  `json:"error,omitempty"`
}

func (h *MonthlyTierReviewHandler) Handle(now time.Time) (MonthlyTierReviewResult, error) {
	startedAt, endedAt := previousMonthRange(now)
	result := MonthlyTierReviewResult{
		Month:     startedAt.Format("2006-01"),
		StartedAt: startedAt,
		EndedAt:   endedAt,
	}

	var partners []database.Partner
	if err := h.DB.Preload("Tier").
		Where("status = ?", database.PartnerStatusActive).
		Find(&partners).Error; err != nil {
		return result, err
	}

	var tiers []database.PartnerTier
	if err := h.DB.Order("level ASC").Find(&tiers).Error; err != nil {
		return result, err
	}

	result.Total = len(partners)
	for _, partner := range partners {
		item := MonthlyTierReviewItem{
			PartnerID:    partner.ID,
			PreviousTier: partner.TierID,
		}

		totalVolume, activeClients, err := h.partnerMonthlyMetrics(partner.ID, startedAt, endedAt)
		if err != nil {
			result.Failed++
			item.Error = err.Error()
			result.PartnerStats = append(result.PartnerStats, item)
			continue
		}
		item.TotalVolume = totalVolume
		item.ActiveClients = activeClients

		bestTier := bestTierForMetrics(tiers, totalVolume, activeClients)
		if bestTier == nil {
			result.Unchanged++
			item.AssignedTier = partner.TierID
			result.PartnerStats = append(result.PartnerStats, item)
			continue
		}
		item.AssignedTier = bestTier.ID
		if bestTier.ID == partner.TierID {
			result.Unchanged++
			result.PartnerStats = append(result.PartnerStats, item)
			continue
		}

		if err := h.DB.Model(&database.Partner{}).
			Where("id = ?", partner.ID).
			Update("tier_id", bestTier.ID).Error; err != nil {
			result.Failed++
			item.Error = err.Error()
			result.PartnerStats = append(result.PartnerStats, item)
			continue
		}
		result.Updated++
		result.PartnerStats = append(result.PartnerStats, item)
	}

	return result, nil
}

func (h *MonthlyTierReviewHandler) partnerMonthlyMetrics(partnerID string, startedAt time.Time, endedAt time.Time) (float64, int64, error) {
	var totalVolume float64
	if err := h.DB.Model(&database.Commission{}).
		Where("partner_id = ?", partnerID).
		Where("trade_date >= ? AND trade_date < ?", startedAt, endedAt).
		Select("COALESCE(SUM(volume_usd), 0)").
		Scan(&totalVolume).Error; err != nil {
		return 0, 0, err
	}

	var activeClients int64
	if err := h.DB.Model(&database.Commission{}).
		Where("partner_id = ?", partnerID).
		Where("trade_date >= ? AND trade_date < ?", startedAt, endedAt).
		Distinct("referred_user_id").
		Count(&activeClients).Error; err != nil {
		return 0, 0, err
	}

	return totalVolume, activeClients, nil
}
