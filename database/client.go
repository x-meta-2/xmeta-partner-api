package database

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	db, err := gorm.Open(postgres.Open(fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=disable TimeZone=%s",
		viper.GetString("DB_HOST"),
		viper.GetInt("DB_PORT"),
		viper.GetString("DB_USER"),
		viper.GetString("DB_NAME"),
		viper.GetString("DB_PASSWORD"),
		viper.GetString("DB_TIMEZONE"),
	)), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		fmt.Println("DATABASE ERROR:", err.Error())
		panic(err.Error())
	}

	fmt.Println("DATABASE CONNECTED...")

	if viper.GetBool("DB_AUTO_MIGRATE") {
		log.Println("DB_AUTO_MIGRATE=true, running AutoMigrate...")
		if err := db.AutoMigrate(
			// Partner core
			&Partner{},
			&PartnerApplication{},
			&PartnerTier{},

			// Referrals
			&ReferralLink{},
			&Referral{},

			// Commissions
			&Commission{},

			// Payouts
			&Payout{},
			&PayoutItem{},

			// NOTE: admin_permissions, admin_groups, admin_group_permissions,
			// admin_users tables are managed by xmeta-admin-api (shared DB).
			// Do NOT auto-migrate them here.
			//
			// NOTE: partner_notifications, partner_activity_logs,
			// partner_daily_stats removed — add back when features are ready.
		); err != nil {
			panic(err.Error())
		}
	} else {
		log.Println("DB_AUTO_MIGRATE=false, skipping AutoMigrate")
	}

	// Custom migrations always run — each step is idempotent. Use this
	// for changes AutoMigrate can't handle safely (drop columns, drop
	// uniques, partial indexes, data backfills, …). Toggle individual
	// migrations on/off inside RunMigrations.
	if viper.GetBool("DB_RUN_MIGRATIONS") {
		RunMigrations(db)
	} else {
		log.Println("DB_RUN_MIGRATIONS=false, skipping custom migrations")
	}

	return db
}
