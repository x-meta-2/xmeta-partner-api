package queries

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/referral/app/dto"
	"xmeta-partner/internal/referral/domain"
	"xmeta-partner/internal/referral/port"

	"gorm.io/gorm"
)

type AdminReferralDetailHandler struct {
	DB        *gorm.DB
	Referrals port.ReferralRepo
}

func (h *AdminReferralDetailHandler) Handle(id string) (*dto.AdminReferralDetail, error) {
	referral, err := h.Referrals.FindByID(id)
	if err != nil {
		return nil, domain.ErrReferralNotFound
	}

	var commissions []database.Commission
	h.DB.
		Where("partner_id = ? AND referred_user_id = ?", referral.PartnerID, referral.ReferredUserID).
		Order("trade_date desc").
		Limit(50).
		Find(&commissions)

	var totalEarned float64
	h.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND referred_user_id = ?", referral.PartnerID, referral.ReferredUserID).
		Select("COALESCE(SUM(rebate_amount), 0)").
		Scan(&totalEarned)

	var totalVolume float64
	h.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND referred_user_id = ?", referral.PartnerID, referral.ReferredUserID).
		Select("COALESCE(SUM(volume_usd), 0)").
		Scan(&totalVolume)

	return &dto.AdminReferralDetail{
		Referral:    *referral,
		Commissions: commissions,
		TotalEarned: totalEarned,
		TotalVolume: totalVolume,
	}, nil
}
