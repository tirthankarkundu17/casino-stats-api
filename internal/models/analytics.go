package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GGRResult represents the GGR per currency
type GGRResult struct {
	Currency  string               `bson:"_id" json:"currency"`
	TotalGGR  primitive.Decimal128 `bson:"totalGGR" json:"totalGGR"`
	USDAmount primitive.Decimal128 `bson:"usdAmount" json:"usdAmount"`
}

// DailyVolumeResult represents volume per day per currency
type DailyVolumeResult struct {
	ID struct {
		Date     time.Time `bson:"date" json:"-"`
		Currency string    `bson:"currency" json:"currency"`
	} `bson:"_id" json:"-"`
	Date        string               `json:"date"`
	Currency    string               `json:"currency"`
	TotalVolume primitive.Decimal128 `bson:"totalVolume" json:"totalVolume"`
	USDVolume   primitive.Decimal128 `bson:"usdVolume" json:"usdVolume"`
}

// UserWagerRank represents the wager rank of a user
type UserWagerRank struct {
	UserID   primitive.ObjectID   `bson:"_id"`
	TotalUSD primitive.Decimal128 `bson:"totalUSD"`
}

// WagerPercentileResult represents the wager percentile of a user
type WagerPercentileResult struct {
	UserID          string  `json:"userId"`
	TotalUSDWagered float64 `json:"totalUsdWagered"`
	Percentile      float64 `json:"percentile"`
}

// UserWagerStats represents the wager stats of a user
type UserWagerStats struct {
	UserID     primitive.ObjectID   `bson:"_id"`
	TotalUSD   primitive.Decimal128 `bson:"totalUSD"`
	Rank       int64                `bson:"rank"`
	TotalUsers int64                `bson:"totalUsers"`
}
