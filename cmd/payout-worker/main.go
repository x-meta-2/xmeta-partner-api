package main

import (
	"log"
	"os"

	"xmeta-partner/database"
	internalPayout "xmeta-partner/internal/payout"
	"xmeta-partner/utils"
)

func main() {
	log.Println("[PayoutWorker] Initializing...")

	utils.LoadConfig()
	db := database.Connect()

	svc := internalPayout.NewService(db)

	if err := svc.Commands.ProcessDailyPayouts.Handle(); err != nil {
		log.Printf("[PayoutWorker] Error: %v", err)
		os.Exit(1)
	}

	log.Println("[PayoutWorker] Done.")
}
