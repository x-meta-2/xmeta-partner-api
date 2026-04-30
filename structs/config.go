package structs

type TierCreateParams struct {
	Name             string   `json:"name" binding:"required"`
	Level            int      `json:"level" binding:"required"`
	CommissionRate   float64  `json:"commissionRate" binding:"required"`
	MinActiveClients int      `json:"minActiveClients"`
	MinVolume        float64  `json:"minVolume"`
	MaxVolume        *float64 `json:"maxVolume"`
	IsDefault        bool     `json:"isDefault"`
	Color            string   `json:"color"`
}

type TierUpdateParams struct {
	Name             *string  `json:"name"`
	Level            *int     `json:"level"`
	CommissionRate   *float64 `json:"commissionRate"`
	MinActiveClients *int     `json:"minActiveClients"`
	MinVolume        *float64 `json:"minVolume"`
	MaxVolume        *float64 `json:"maxVolume"`
	IsDefault        *bool    `json:"isDefault"`
	Color            *string  `json:"color"`
}
