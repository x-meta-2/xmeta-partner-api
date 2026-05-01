package partner

import (
	"fmt"
	"time"

	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/services"
	"xmeta-partner/structs"
	"xmeta-partner/utils"

	"gorm.io/gorm"
)

type ReferralService struct {
	services.BaseService
}

// ReferralUserRef is the partner-safe view of a referred user. Drops
// every field a partner has no business seeing — full email, internal
// Binance sub-account, capability flags, metadata, etc. — and exposes
// only what a partner dashboard legitimately needs to display.
type ReferralUserRef struct {
	ID          string `json:"id"`
	MaskedEmail string `json:"maskedEmail"`
	FirstName   string `json:"firstName"`
	LastInitial string `json:"lastInitial"`
	KycLevel    int    `json:"kycLevel"`
}

// ReferralListItem mirrors database.Referral but with the user object
// scrubbed of PII. Returned by the partner-facing List/Detail endpoints.
type ReferralListItem struct {
	ID             string                `json:"id"`
	PartnerID      string                `json:"partnerId"`
	ReferredUserID string                `json:"referredUserId"`
	ReferredUser   *ReferralUserRef      `json:"referredUser"`
	ReferralLinkID *string               `json:"referralLinkId"`
	Status         database.ReferralStatus `json:"status"`
	StartedAt      time.Time             `json:"startedAt"`
	EndedAt        *time.Time            `json:"endedAt"`
	RegisteredAt   time.Time             `json:"registeredAt"`
	FirstDepositAt *time.Time            `json:"firstDepositAt"`
	FirstTradeAt   *time.Time            `json:"firstTradeAt"`
	CreatedAt      time.Time             `json:"createdAt"`
}

func toReferralListItem(r database.Referral) ReferralListItem {
	item := ReferralListItem{
		ID:             r.ID,
		PartnerID:      r.PartnerID,
		ReferredUserID: r.ReferredUserID,
		ReferralLinkID: r.ReferralLinkID,
		Status:         r.Status,
		StartedAt:      r.StartedAt,
		EndedAt:        r.EndedAt,
		RegisteredAt:   r.RegisteredAt,
		FirstDepositAt: r.FirstDepositAt,
		FirstTradeAt:   r.FirstTradeAt,
		CreatedAt:      r.CreatedAt,
	}
	if r.ReferredUser != nil {
		item.ReferredUser = &ReferralUserRef{
			ID:          r.ReferredUser.ID,
			MaskedEmail: utils.MaskEmail(r.ReferredUser.Email),
			FirstName:   r.ReferredUser.FirstName,
			LastInitial: utils.LastInitial(r.ReferredUser.LastName),
			KycLevel:    r.ReferredUser.KycLevel,
		}
	}
	return item
}

func (s *ReferralService) List(partnerID string, params structs.ReferralListParams) (structs.PaginationResponse, error) {
	pInput := services.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	// Hide historical/unlinked rows from the partner-facing list. Those
	// stay in DB for commission attribution but the partner only cares
	// about people who are currently linked to them.
	orm := s.DB.Model(&database.Referral{}).
		Where("partner_id = ? AND ended_at IS NULL", partnerID)
	orm = common.Equal(orm, "status", params.Status)

	if params.Query != "" {
		q := "%" + params.Query + "%"
		orm = orm.
			Joins("LEFT JOIN users ON users.id = referrals.referred_user_id").
			Where("users.email ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ?", q, q, q)
	}

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var referrals []database.Referral
	if err := orm.
		Preload("ReferredUser").
		Order("created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&referrals).Error; err != nil {
		return structs.PaginationResponse{}, err
	}

	items := make([]ReferralListItem, len(referrals))
	for i := range referrals {
		items[i] = toReferralListItem(referrals[i])
	}
	return structs.PaginationResponse{Total: total, Items: items}, nil
}

func (s *ReferralService) Detail(partnerID string, id string) (ReferralListItem, error) {
	var referral database.Referral
	if err := s.DB.Preload("ReferredUser").
		Where("id = ? AND partner_id = ?", id, partnerID).
		First(&referral).Error; err != nil {
		return ReferralListItem{}, err
	}
	return toReferralListItem(referral), nil
}

// Stats counts the partner's *currently linked* referrals only — historical
// rows (ended_at IS NOT NULL) are filtered out so users that switched away
// don't keep inflating the partner's totals.
func (s *ReferralService) Stats(partnerID string) (map[string]interface{}, error) {
	base := func() *gorm.DB {
		return s.DB.Model(&database.Referral{}).
			Where("partner_id = ? AND ended_at IS NULL", partnerID)
	}

	var totalCount, registeredCount, depositedCount, activeCount, inactiveCount int64
	base().Count(&totalCount)
	base().Where("status = ?", "registered").Count(&registeredCount)
	base().Where("status = ?", "deposited").Count(&depositedCount)
	base().Where("status = ?", "active").Count(&activeCount)
	base().Where("status = ?", "inactive").Count(&inactiveCount)

	return map[string]interface{}{
		"total":      totalCount,
		"registered": registeredCount,
		"deposited":  depositedCount,
		"active":     activeCount,
		"inactive":   inactiveCount,
	}, nil
}

// ProcessUserRegistered handles a fresh signup arriving via the monorepo
// `/internal/user-registered` hook. The user is brand-new (no prior
// referral row), so this is a straight Link.
func (s *ReferralService) ProcessUserRegistered(params structs.UserRegisteredParams) error {
	return s.LinkReferral(params.UserID, params.ReferralCode)
}

// LinkReferral hooks an xmeta user under a partner's code. Called by the
// partner-portal Settings page and by ProcessUserRegistered at signup
// time. If the user is currently linked to another partner, that
// relationship is closed first and a new active row is opened —
// commissions table holds the immutable per-trade attribution, so the
// switch is safe.
func (s *ReferralService) LinkReferral(userID, code string) error {
	var link database.ReferralLink
	if err := s.DB.Where("code = ? AND is_active = ?", code, true).First(&link).Error; err != nil {
		return fmt.Errorf("referral link not found: %s", code)
	}

	// Verify the partner is active and not the user themselves.
	var partner database.Partner
	if err := s.DB.Where("id = ?", link.PartnerID).First(&partner).Error; err != nil {
		return fmt.Errorf("partner not found for code: %s", code)
	}
	if partner.Status != database.PartnerStatusActive {
		return fmt.Errorf("referral code is not active")
	}
	if partner.UserID == userID {
		return fmt.Errorf("cannot link to your own referral code")
	}

	now := time.Now()

	return s.DB.Transaction(func(tx *gorm.DB) error {
		// Close any currently active referral so the partial unique index
		// `WHERE ended_at IS NULL` lets the new row through.
		if err := tx.Model(&database.Referral{}).
			Where("referred_user_id = ? AND ended_at IS NULL", userID).
			Updates(map[string]interface{}{
				"ended_at": now,
				"status":   database.ReferralStatusUnlinked,
			}).Error; err != nil {
			return err
		}

		referral := database.Referral{
			PartnerID:      link.PartnerID,
			ReferredUserID: userID,
			ReferralLinkID: &link.ID,
			Status:         database.ReferralStatusRegistered,
			StartedAt:      now,
			RegisteredAt:   now,
		}
		if err := tx.Create(&referral).Error; err != nil {
			return err
		}

		// Counters bump only on first-ever link (switch-flow doesn't re-bump).
		var historyCount int64
		tx.Model(&database.Referral{}).Where("referred_user_id = ?", userID).Count(&historyCount)
		if historyCount == 1 {
			tx.Model(&link).UpdateColumn("registrations", gorm.Expr("registrations + 1"))
			tx.Model(&database.Partner{}).Where("id = ?", link.PartnerID).
				UpdateColumn("total_referrals", gorm.Expr("total_referrals + 1"))
		}

		return nil
	})
}

// ReferralLinkLookup is the response for `GET /internal/referral-links/:code`.
// Only enough partner identity is included to surface "this code belongs
// to {Name}" feedback as the monorepo user types a code into the form.
type ReferralLinkLookup struct {
	Code            string `json:"code"`
	IsActive        bool   `json:"isActive"`
	PartnerID       string `json:"partnerId"`
	PartnerEmail    string `json:"partnerEmail,omitempty"`
	PartnerFullName string `json:"partnerFullName,omitempty"`
}

// LookupReferralLink validates a referral code and returns the partner's
// public-safe identity. Returns an error if the code is unknown.
//
// `isActive` answers "can this code attribute new referrals right now" —
// it requires both the link itself to be active AND the partner to be in
// status=active. A suspended partner's code reads back as `isActive=false`
// even if `referral_links.is_active=true`, so the monorepo settings form
// can warn the user before they submit.
func (s *ReferralService) LookupReferralLink(code string) (ReferralLinkLookup, error) {
	var link database.ReferralLink
	if err := s.DB.Where("code = ?", code).First(&link).Error; err != nil {
		return ReferralLinkLookup{}, fmt.Errorf("referral code not found")
	}

	out := ReferralLinkLookup{
		Code:      link.Code,
		PartnerID: link.PartnerID,
	}

	var partner database.Partner
	if err := s.DB.Preload("User").Where("id = ?", link.PartnerID).First(&partner).Error; err == nil {
		out.IsActive = link.IsActive && partner.Status == database.PartnerStatusActive
		if partner.User != nil {
			out.PartnerEmail = partner.User.Email
			out.PartnerFullName = partner.User.FirstName + " " + partner.User.LastName
		}
	}

	return out, nil
}

// UnlinkReferral closes the user's currently active referral. Past
// commissions stay attached to whichever partner was active at the time
// of each trade — commissions.partner_id is captured at insert and is
// never rewritten.
func (s *ReferralService) UnlinkReferral(userID string) error {
	now := time.Now()
	result := s.DB.Model(&database.Referral{}).
		Where("referred_user_id = ? AND ended_at IS NULL", userID).
		Updates(map[string]interface{}{
			"ended_at": now,
			"status":   database.ReferralStatusUnlinked,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no active referral to unlink")
	}
	return nil
}

// ProcessUserDeposited records the user's first deposit on their currently
// active referral row. If the user has already moved past "deposited"
// (i.e. they've traded → status=active/inactive), only `first_deposit_at`
// is set so we don't regress the lifecycle.
func (s *ReferralService) ProcessUserDeposited(params structs.UserDepositedParams) error {
	var referral database.Referral
	if err := s.DB.Where("referred_user_id = ? AND ended_at IS NULL", params.UserID).
		First(&referral).Error; err != nil {
		return fmt.Errorf("no active referral for user: %s", params.UserID)
	}
	if referral.FirstDepositAt != nil {
		return nil // idempotent — already recorded
	}

	now := time.Now()
	updates := map[string]interface{}{"first_deposit_at": now}
	// Only advance the status if the lifecycle is still pre-deposit.
	if referral.Status == database.ReferralStatusRegistered {
		updates["status"] = database.ReferralStatusDeposited
	}

	return s.DB.Model(&referral).Updates(updates).Error
}
