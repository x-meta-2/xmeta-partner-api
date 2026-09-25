package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"xmeta-partner/database"
	internalCommission "xmeta-partner/internal/commission"
	"xmeta-partner/structs"
	"xmeta-partner/utils"
)

const (
	defaultSyncLimit        = 10000
	defaultSyncLookbackDays = 3
)

func main() {
	log.Println("[FuturesCommissionSync] Initializing...")

	utils.LoadConfig()
	db := database.Connect()
	svc := internalCommission.NewService(db)

	params := syncParamsFromEnv()
	log.Printf(
		"[FuturesCommissionSync] startedAt=%s endedAt=%s limit=%d",
		params.StartedAt,
		params.EndedAt,
		params.Limit,
	)

	result, err := svc.Commands.SyncFuturesClosedPositions.Handle(params)
	if err != nil {
		log.Printf("[FuturesCommissionSync] Error: %v", err)
		os.Exit(1)
	}

	log.Printf(
		"[FuturesCommissionSync] Done total=%d success=%d skipped=%d failed=%d",
		result.Total,
		result.Success,
		result.Skipped,
		result.Failed,
	)
	if result.Failed > 0 {
		for _, item := range result.Errors {
			log.Printf(
				"[FuturesCommissionSync] positionId=%s userId=%s message=%s",
				item.PositionID,
				item.UserID,
				item.Message,
			)
		}
		os.Exit(1)
	}
}

func syncParamsFromEnv() structs.FuturesClosedPositionSyncParams {
	startedAt := os.Getenv("FUTURES_SYNC_STARTED_AT")
	endedAt := os.Getenv("FUTURES_SYNC_ENDED_AT")
	if startedAt == "" || endedAt == "" {
		location, err := time.LoadLocation("Asia/Ulaanbaatar")
		if err != nil {
			location = time.Local
		}
		now := time.Now().In(location)
		lookbackDays := defaultSyncLookbackDays
		if raw := os.Getenv("FUTURES_SYNC_LOOKBACK_DAYS"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				lookbackDays = parsed
			}
		}
		yesterday := now.AddDate(0, 0, -1)
		if startedAt == "" {
			startedAt = yesterday.AddDate(0, 0, -(lookbackDays - 1)).Format("2006-01-02")
		}
		if endedAt == "" {
			endedAt = yesterday.Format("2006-01-02")
		}
	}

	limit := defaultSyncLimit
	if raw := os.Getenv("FUTURES_SYNC_LIMIT"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	return structs.FuturesClosedPositionSyncParams{
		StartedAt: startedAt,
		EndedAt:   endedAt,
		Limit:     limit,
	}
}
