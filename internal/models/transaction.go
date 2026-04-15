package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Transaction represents a casino game transaction (Wager or Payout)
type Transaction struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time            `bson:"createdAt" json:"createdAt"`
	UserID    primitive.ObjectID   `bson:"userId" json:"userId"`
	RoundID   string               `bson:"roundId" json:"roundId"`
	Type      string               `bson:"type" json:"type"` // "Wager" or "Payout"
	Amount    primitive.Decimal128 `bson:"amount" json:"amount"`
	Currency  string               `bson:"currency" json:"currency"`   // "ETH", "BTC", "USDT"
	USDAmount primitive.Decimal128 `bson:"usdAmount" json:"usdAmount"` // USD value of the transaction
}

// Request models for validation
type TimeframeParams struct {
	From time.Time `form:"from" binding:"required" time_format:"2006-01-02"`
	To   time.Time `form:"to" binding:"required" time_format:"2006-01-02"`
}

// Validate checks that the timeframe is logically valid.
func (p *TimeframeParams) Validate() error {
	if !p.From.Before(p.To) {
		return fmt.Errorf("'from' date must be before 'to' date")
	}
	if p.From.After(time.Now()) {
		return fmt.Errorf("'from' date cannot be in the future")
	}
	return nil
}

// UserWagerPercentileParams represents the wager percentile of a user
type UserWagerPercentileParams struct {
	TimeframeParams
	UserID string `uri:"user_id" binding:"required"`
}
