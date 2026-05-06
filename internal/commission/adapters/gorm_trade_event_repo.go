package adapters

import (
	"time"

	"xmeta-partner/database"

	"gorm.io/gorm"
)

type GormTradeEventRepo struct {
	DB *gorm.DB
}

func (r *GormTradeEventRepo) ExistsByPositionID(positionID string) (bool, error) {
	var count int64
	if err := r.DB.Model(&database.Commission{}).Where("position_id = ?", positionID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormTradeEventRepo) IsUserKycVerified(userID string) (bool, error) {
	var user database.User
	if err := r.DB.Select("kyc_level").Where("id = ?", userID).First(&user).Error; err != nil {
		return false, err
	}
	return user.KycLevel >= 1, nil
}

func (r *GormTradeEventRepo) FindActiveReferral(userID string, tradeDate time.Time) (*database.Referral, error) {
	var referral database.Referral
	err := r.DB.
		Where("referred_user_id = ? AND started_at <= ? AND (ended_at IS NULL OR ended_at > ?)", userID, tradeDate, tradeDate).
		First(&referral).Error
	if err != nil {
		return nil, err
	}
	return &referral, nil
}

func (r *GormTradeEventRepo) FindActivePartnerWithTier(partnerID string) (*database.Partner, error) {
	var partner database.Partner
	err := r.DB.Preload("Tier").Where("id = ? AND status = ?", partnerID, database.PartnerStatusActive).First(&partner).Error
	if err != nil {
		return nil, err
	}
	return &partner, nil
}

func (r *GormTradeEventRepo) CreateCommission(c *database.Commission) error {
	return r.DB.Create(c).Error
}

func (r *GormTradeEventRepo) IncrementPartnerEarnings(partnerID string, amount float64) error {
	return r.DB.Model(&database.Partner{}).
		Where("id = ?", partnerID).
		Update("total_earnings", gorm.Expr("total_earnings + ?", amount)).Error
}

func (r *GormTradeEventRepo) ActivateReferral(referralID string, firstTradeAt time.Time) error {
	return r.DB.Model(&database.Referral{}).
		Where("id = ? AND first_trade_at IS NULL", referralID).
		Updates(map[string]any{
			"first_trade_at": &firstTradeAt,
			"status":         database.ReferralStatusActive,
		}).Error
}
