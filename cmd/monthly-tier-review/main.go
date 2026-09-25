package main

import (
	"log"
	"os"
	"time"

	"xmeta-partner/database"
	internalCommission "xmeta-partner/internal/commission"
	"xmeta-partner/utils"
)

func main() {
	log.Println("[MonthlyTierReview] Initializing...")

	utils.LoadConfig()
	db := database.Connect()
	svc := internalCommission.NewService(db)

	result, err := svc.Commands.MonthlyTierReview.Handle(time.Now())
	if err != nil {
		log.Printf("[MonthlyTierReview] Error: %v", err)
		os.Exit(1)
	}

	log.Printf(
		"[MonthlyTierReview] Done month=%s total=%d updated=%d unchanged=%d failed=%d",
		result.Month,
		result.Total,
		result.Updated,
		result.Unchanged,
		result.Failed,
	)
	for _, item := range result.PartnerStats {
		if item.Error != "" {
			log.Printf(
				"[MonthlyTierReview] partnerId=%s previousTier=%s assignedTier=%s volume=%.4f activeClients=%d error=%s",
				item.PartnerID,
				item.PreviousTier,
				item.AssignedTier,
				item.TotalVolume,
				item.ActiveClients,
				item.Error,
			)
			continue
		}
		log.Printf(
			"[MonthlyTierReview] partnerId=%s previousTier=%s assignedTier=%s volume=%.4f activeClients=%d",
			item.PartnerID,
			item.PreviousTier,
			item.AssignedTier,
			item.TotalVolume,
			item.ActiveClients,
		)
	}

	if result.Failed > 0 {
		os.Exit(1)
	}
}
