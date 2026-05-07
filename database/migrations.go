package database

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// RunMigrations applies one-off schema changes that AutoMigrate can't make
// safely on its own (drop columns, drop unique constraints, add partial
// indexes, backfill data, etc.). Each step is idempotent — feel free to
// leave the call in place after it has run; subsequent boots no-op.
//
// Standard flow when you need to change a table:
//  1. Add the helper / migration block below.
//  2. Bring the server up — migration runs once, log the result.
//  3. After the change has rolled out everywhere, comment the call out
//     to keep startup fast (the helper itself can stay as a record).
func RunMigrations(db *gorm.DB) {
	log.Println("!!! RunMigrations STARTING !!!")

	// ── Active migrations ──────────────────────────────────────────────
	migrateReferralsToTimeBound(db)    // 2026-04-29 — switchable referrals
	dropDeadReferralColumns(db)        // 2026-04-29 — utm + bonus + ip/ua were never wired
	dropSubAffiliateArtifacts(db)      // 2026-04-30 — full sub-affiliate removal
	migrateCommissionsSchema(db)       // 2026-05-05 — align with monorepo trade event format
	renameCommissionColumns(db)        // 2026-05-06 — fee_amount → commission_amount, commission_amount → rebate_amount
	ensureDefaultTier(db)              // 2026-05-07 — guarantee at least one default tier exists
	addPayoutConcurrencyGuard(db)      // 2026-05-07 — partial unique index: one pending/processing payout per partner

	log.Println("Custom migrations completed!")
}

// migrateReferralsToTimeBound rewrites the `referrals` table so a single
// user can have a *history* of partner relationships:
//
//   - drops the strict UNIQUE index on `referred_user_id`
//   - adds `started_at` (NOT NULL) and `ended_at` (nullable) timestamps
//   - backfills `started_at` from `created_at` for existing rows
//   - creates a *partial* unique index that keeps "one active referral
//     per user" while allowing as many disconnected/historical rows as
//     the partner-switch flow needs
//
// The commission engine queries by trade_date inside the active window so
// past trades stay attributed to the partner who was active at the time.
func migrateReferralsToTimeBound(db *gorm.DB) {
	log.Println("→ migrateReferralsToTimeBound")

	// 1. Drop the strict 1:1 unique constraint (any of the names GORM
	//    might have generated). DROP IF EXISTS is idempotent.
	for _, idx := range []string{
		"idx_referrals_referred_user_id",
		"uni_referrals_referred_user_id",
		"referrals_referred_user_id_key",
	} {
		dropIndex(db, idx)
	}

	// 2. Timestamps
	addColumnIfMissing(db, "referrals", "started_at", "TIMESTAMPTZ")
	addColumnIfMissing(db, "referrals", "ended_at", "TIMESTAMPTZ")

	// 3. Backfill existing rows so `started_at` always has a value before
	//    we set NOT NULL.
	if err := db.Exec(`
		UPDATE referrals
		SET started_at = created_at
		WHERE started_at IS NULL
	`).Error; err != nil {
		log.Printf("  Error backfilling referrals.started_at: %v", err)
		return
	}

	// 4. Tighten the column once it's safe.
	if err := db.Exec(`
		ALTER TABLE referrals ALTER COLUMN started_at SET NOT NULL
	`).Error; err != nil {
		log.Printf("  Error setting NOT NULL on referrals.started_at: %v", err)
	}

	// 5. Partial unique index: at most one *active* referral per user.
	//    Allows multiple historical (ended_at IS NOT NULL) rows alongside.
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS referrals_one_active_per_user
		ON referrals (referred_user_id)
		WHERE ended_at IS NULL
	`).Error; err != nil {
		log.Printf("  Error creating partial unique index: %v", err)
	}

	log.Println("✓ referrals table is now time-bound")
}

// dropSubAffiliateArtifacts removes every DB artifact left over from the
// (now-removed) 2-tier affiliate feature:
//
//   - sub_affiliate_invites table (whole feature was never wired up)
//   - partners.parent_id      — was meant to chain partners; nothing read it
//   - commissions.is_override — override row marker; no overrides ever ran
//   - commissions.override_partner_id — same
//
// Idempotent — uses `DROP TABLE/COLUMN IF EXISTS` so re-runs are no-ops.
func dropSubAffiliateArtifacts(db *gorm.DB) {
	log.Println("→ dropSubAffiliateArtifacts")
	dropTable(db, "sub_affiliate_invites")
	dropColumn(db, "partners", "parent_id")
	dropColumn(db, "commissions", "is_override")
	dropColumn(db, "commissions", "override_partner_id")
}

// dropDeadReferralColumns removes the audit/bonus columns that the live
// flow never populates: ip_address, user_agent, utm_source, utm_medium,
// utm_campaign, signup_bonus, referee_bonus. Idempotent — uses
// `DROP COLUMN IF EXISTS` so a second boot is a no-op.
func dropDeadReferralColumns(db *gorm.DB) {
	log.Println("→ dropDeadReferralColumns")
	for _, col := range []string{
		"ip_address",
		"user_agent",
		"utm_source",
		"utm_medium",
		"utm_campaign",
		"signup_bonus",
		"referee_bonus",
	} {
		dropColumn(db, "referrals", col)
	}
}

// migrateCommissionsSchema aligns the commissions table with the monorepo
// trade event format:
//
//   - trade_id      → position_id
//   - trade_amount  → fee_amount (via trade_fee intermediate)
//   - trade_fee     → fee_amount
//   - commission_asset → asset
//   - NEW market_id, asset, fee_amount, volume_usd
//   - DROP asset_id (was composite key experiment)
func migrateCommissionsSchema(db *gorm.DB) {
	log.Println("→ migrateCommissionsSchema")

	addColumnIfMissing(db, "commissions", "market_id", "VARCHAR NOT NULL DEFAULT ''")
	addColumnIfMissing(db, "commissions", "volume_usd", "DECIMAL(20,8) NOT NULL DEFAULT 0")

	if hasColumn(db, "commissions", "trade_id") && hasColumn(db, "commissions", "position_id") {
		db.Exec("UPDATE commissions SET position_id = trade_id WHERE position_id = ''")
	}
	if hasColumn(db, "commissions", "trade_fee") && hasColumn(db, "commissions", "fee_amount") {
		db.Exec("UPDATE commissions SET fee_amount = trade_fee WHERE fee_amount = 0")
	}
	if hasColumn(db, "commissions", "trade_amount") && hasColumn(db, "commissions", "fee_amount") {
		db.Exec("UPDATE commissions SET fee_amount = trade_amount WHERE fee_amount = 0")
	}
	if hasColumn(db, "commissions", "commission_asset") && hasColumn(db, "commissions", "asset") {
		db.Exec("UPDATE commissions SET asset = commission_asset WHERE asset = ''")
	}

	dropColumn(db, "commissions", "trade_id")
	dropColumn(db, "commissions", "trade_amount")
	dropColumn(db, "commissions", "trade_fee")
	dropColumn(db, "commissions", "commission_asset")
	dropColumn(db, "commissions", "asset_id")

	for _, idx := range []string{
		"idx_position_asset",
		"idx_commissions_trade_id",
	} {
		dropIndex(db, idx)
	}

	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_commissions_position_id ON commissions(position_id)")

	log.Println("✓ commissions schema aligned with monorepo format")
}

// renameCommissionColumns renames the Commission model columns to clarify
// semantics: fee_amount → commission_amount (the raw fee from monorepo),
// commission_amount → rebate_amount (the partner's share after rate applied).
// Order matters: rename commission_amount first so the name is free for
// the second rename.
func renameCommissionColumns(db *gorm.DB) {
	log.Println("→ renameCommissionColumns")
	renameColumn(db, "commissions", "commission_amount", "rebate_amount")
	renameColumn(db, "commissions", "fee_amount", "commission_amount")
}

func ensureDefaultTier(db *gorm.DB) {
	log.Println("→ ensureDefaultTier")

	var defaults []PartnerTier
	db.Where("is_default = ?", true).Order("created_at asc").Find(&defaults)

	if len(defaults) > 1 {
		log.Printf("  Found %d default tiers — keeping oldest, clearing rest", len(defaults))
		keep := defaults[0].ID
		db.Model(&PartnerTier{}).
			Where("is_default = ? AND id != ?", true, keep).
			Update("is_default", false)
	}

	if len(defaults) == 0 {
		log.Println("  No default tier found, creating Standard tier")
		tier := PartnerTier{
			Name:           "Standard",
			Level:          1,
			CommissionRate: 0.20,
			IsDefault:      true,
			Color:          "#6b7280",
		}
		if err := db.Create(&tier).Error; err != nil {
			log.Printf("  Error creating default tier: %v", err)
			return
		}
	}

	db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_partner_tiers_one_default
		ON partner_tiers (is_default)
		WHERE is_default = true
	`)

	log.Println("✓ Default tier invariant enforced")
}

// addPayoutConcurrencyGuard creates a partial unique index on the payouts
// table so that each partner can have at most one active (pending or
// processing) payout at any time. The advisory lock in the application
// layer is the first line of defence; this index is the DB-level backstop.
func addPayoutConcurrencyGuard(db *gorm.DB) {
	log.Println("→ addPayoutConcurrencyGuard")

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_payouts_one_active_per_partner
		ON payouts (partner_id)
		WHERE status IN ('pending', 'processing')
	`).Error; err != nil {
		log.Printf("  Error creating payout concurrency index: %v", err)
		return
	}

	log.Println("✓ Payout concurrency guard index created")
}

// ─── helpers ──────────────────────────────────────────────────────────

func hasColumn(db *gorm.DB, table, column string) bool {
	return db.Migrator().HasColumn(table, column)
}

func addColumnIfMissing(db *gorm.DB, table, column, dataType string) {
	if hasColumn(db, table, column) {
		return
	}
	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, dataType)
	if err := db.Exec(sql).Error; err != nil {
		log.Printf("  Error adding column %s.%s: %v", table, column, err)
		return
	}
	log.Printf("  ✓ Added column %s.%s", table, column)
}

func dropColumn(db *gorm.DB, table, column string) {
	if !hasColumn(db, table, column) {
		return
	}
	if err := db.Migrator().DropColumn(table, column); err != nil {
		log.Printf("  Error dropping column %s.%s: %v", table, column, err)
		return
	}
	log.Printf("  ✓ Dropped column %s.%s", table, column)
}

func dropIndex(db *gorm.DB, indexName string) {
	if err := db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", indexName)).Error; err != nil {
		log.Printf("  Error dropping index %s: %v", indexName, err)
		return
	}
}

func dropTable(db *gorm.DB, table string) {
	if !db.Migrator().HasTable(table) {
		return
	}
	if err := db.Migrator().DropTable(table); err != nil {
		log.Printf("  Error dropping table %s: %v", table, err)
		return
	}
	log.Printf("  ✓ Dropped table %s", table)
}

func renameColumn(db *gorm.DB, table, oldName, newName string) {
	if !hasColumn(db, table, oldName) || hasColumn(db, table, newName) {
		return
	}
	if err := db.Migrator().RenameColumn(table, oldName, newName); err != nil {
		log.Printf("  Error renaming %s.%s → %s: %v", table, oldName, newName, err)
		return
	}
	log.Printf("  ✓ Renamed %s.%s → %s", table, oldName, newName)
}
