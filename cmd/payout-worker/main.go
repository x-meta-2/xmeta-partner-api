package main

import (
	"log"
	"os"

	"xmeta-partner/database"
	"xmeta-partner/services"
	"xmeta-partner/utils"
)

func main() {
	log.Println("[PayoutWorker] Initializing...")

	utils.LoadConfig()
	db := database.Connect()

	worker := services.PayoutWorkerService{
		BaseService: services.BaseService{DB: db},
	}

	if err := worker.ProcessDailyPayouts(); err != nil {
		log.Printf("[PayoutWorker] Error: %v", err)
		os.Exit(1)
	}

	log.Println("[PayoutWorker] Done.")
}
