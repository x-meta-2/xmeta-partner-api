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

	// ── Completed (kept commented as a changelog) ──────────────────────
	// migrateReferralsToTimeBound(db) // 2026-04-29

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

func dropForeignKey(db *gorm.DB, table, constraint string) {
	sql := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", table, constraint)
	if err := db.Exec(sql).Error; err != nil {
		log.Printf("  Error dropping FK %s on %s: %v", constraint, table, err)
	}
}

func dropNotNullConstraint(db *gorm.DB, table, column string) {
	if !hasColumn(db, table, column) {
		return
	}
	sql := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL", table, column)
	if err := db.Exec(sql).Error; err != nil {
		log.Printf("  Error dropping NOT NULL on %s.%s: %v", table, column, err)
		return
	}
	log.Printf("  ✓ Dropped NOT NULL %s.%s", table, column)
}
