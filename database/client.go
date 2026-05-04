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

		); err != nil {
			panic(err.Error())
		}
	} else {
		log.Println("DB_AUTO_MIGRATE=false, skipping AutoMigrate")
	}
	if viper.GetBool("DB_RUN_MIGRATIONS") {
		RunMigrations(db)
	} else {
		log.Println("DB_RUN_MIGRATIONS=false, skipping custom migrations")
	}

	return db
}
