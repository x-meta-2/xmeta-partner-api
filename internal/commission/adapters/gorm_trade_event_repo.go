package adapters

import (
	"errors"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/commission/domain"
	"xmeta-partner/internal/commission/port"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNoActiveReferral
		}
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
	result := r.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "position_id"}},
		DoNothing: true,
	}).Create(c)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrDuplicatePosition
	}
	return nil
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

func (r *GormTradeEventRepo) GetPartnerTotalVolume(partnerID string) (float64, error) {
	var total float64
	err := r.DB.Model(&database.Commission{}).
		Where("partner_id = ?", partnerID).
		Select("COALESCE(SUM(volume_usd), 0)").
		Scan(&total).Error
	return total, err
}

func (r *GormTradeEventRepo) GetPartnerActiveClients(partnerID string) (int64, error) {
	var count int64
	err := r.DB.Model(&database.Referral{}).
		Where("partner_id = ? AND status = ? AND ended_at IS NULL", partnerID, database.ReferralStatusActive).
		Count(&count).Error
	return count, err
}

func (r *GormTradeEventRepo) FindAllTiersAsc() ([]database.PartnerTier, error) {
	var tiers []database.PartnerTier
	err := r.DB.Order("level ASC").Find(&tiers).Error
	return tiers, err
}

func (r *GormTradeEventRepo) UpgradePartnerTier(partnerID string, newTierID string, newLevel int) error {
	return r.DB.Exec(`
		UPDATE partners SET tier_id = ?
		WHERE id = ?
		AND (SELECT level FROM partner_tiers WHERE id = partners.tier_id) < ?
	`, newTierID, partnerID, newLevel).Error
}

func (r *GormTradeEventRepo) RunInTx(fn func(port.TradeEventRepo) error) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		return fn(&GormTradeEventRepo{DB: tx})
	})
}
