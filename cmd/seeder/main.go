package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"tirthankarkundu17/casino-analytics/internal/config"
	"tirthankarkundu17/casino-analytics/internal/db"
	"tirthankarkundu17/casino-analytics/internal/logger"
	"tirthankarkundu17/casino-analytics/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	currencies = []string{"ETH", "BTC", "USDT"}
	rates      = map[string]float64{
		"ETH":  3000.0,
		"BTC":  60000.0,
		"USDT": 1.0,
	}
)

func main() {
	log := logger.New()
	defer log.Sync()

	cfg := config.LoadConfig(log)

	m, err := db.ConnectMongo(cfg.MongoURI, cfg.MongoDB, log)
	if err != nil {
		log.Fatalf("Failed to connect to Mongo: %v", err)
	}
	defer m.Close()

	coll := m.Db.Collection(cfg.TransactionCollection)

	// Drop existing data to prevent duplication on re-runs
	log.Info("Dropping existing collection...")
	if err := coll.Drop(context.Background()); err != nil {
		log.Fatalf("Failed to drop collection: %v", err)
	}

	log.Info("Creating primary index")
	// Round integrity lookups
	_, err = coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{Key: "roundId", Value: 1}},
	})

	userIds := make([]primitive.ObjectID, cfg.SeederNumUsers)
	for i := 0; i < cfg.SeederNumUsers; i++ {
		userIds[i] = primitive.NewObjectID()
	}

	// Log first 10 user IDs for testing/easy access
	sampleCount := 10
	if len(userIds) < sampleCount {
		sampleCount = len(userIds)
	}
	sampleIds := make([]string, sampleCount)
	for i := 0; i < sampleCount; i++ {
		sampleIds[i] = userIds[i].Hex()
	}
	log.Infof("Sample User IDs for testing: %v", sampleIds)

	log.Infof("Starting seeding %d rounds (~%d transactions). Concurrent: %v", cfg.SeederNumRounds, cfg.SeederNumRounds*2, cfg.SeederConcurrent)

	startTime := time.Now()

	if cfg.SeederConcurrent {
		numWorkers := 10

		roundsPerWorker := cfg.SeederNumRounds / numWorkers
		remainderRounds := cfg.SeederNumRounds % numWorkers

		var wg sync.WaitGroup

		for w := range numWorkers {
			wg.Add(1)

			workerRounds := roundsPerWorker
			if w == numWorkers-1 {
				workerRounds += remainderRounds // give remainder to the last worker
			}

			go func(workerID int, totalRounds int) {
				defer wg.Done()
				seedData(workerID, totalRounds, coll, userIds, log, cfg.SeederNumUsers, cfg.SeederBatchSize)
			}(w, workerRounds)
		}

		wg.Wait()
	} else {
		seedData(0, cfg.SeederNumRounds, coll, userIds, log, cfg.SeederNumUsers, cfg.SeederBatchSize)
	}

	// Create Indexes — compound indexes aligned to aggregation query patterns
	log.Info("Creating indexes...")
	_, err = coll.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		// Covers GetDailyWagerVolume and GetAllUsersWagerRank: $match {type, createdAt}
		{Keys: bson.D{{Key: "type", Value: 1}, {Key: "createdAt", Value: 1}, {Key: "currency", Value: 1}, {Key: "usdAmount", Value: 1}, {Key: "amount", Value: 1}}},

		// Covers GetUserWagerPercentile: $match {type, createdAt, userId}
		{Keys: bson.D{{Key: "type", Value: 1}, {Key: "createdAt", Value: 1}, {Key: "usdAmount", Value: 1}, {Key: "userId", Value: 1}}},
	})
	if err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	log.Infof("Seeding completed in %v", time.Since(startTime))
}

func seedData(workerID int, totalRounds int, coll *mongo.Collection, userIds []primitive.ObjectID, log *zap.SugaredLogger, numUsers int, batchSize int) {
	roundsLeft := totalRounds
	for roundsLeft > 0 {
		currentBatchRounds := batchSize
		if roundsLeft < batchSize {
			currentBatchRounds = roundsLeft
		}

		docs := make([]any, 0, currentBatchRounds*2)
		for i := 0; i < currentBatchRounds; i++ {
			userID := userIds[rand.Intn(numUsers)]
			roundID := primitive.NewObjectID().Hex()
			currency := currencies[rand.Intn(len(currencies))]

			// Wager
			wagerTime := time.Now().AddDate(0, 0, -rand.Intn(365)).Add(time.Duration(rand.Intn(24)) * time.Hour)
			wagerAmount := rand.Float64() * 0.1 // Max 0.1 BTC/ETH etc
			if currency == "USDT" {
				wagerAmount *= 1000
			}

			wagerUSD := wagerAmount * rates[currency]

			wager := models.Transaction{
				ID:        primitive.NewObjectID(),
				CreatedAt: wagerTime,
				UserID:    userID,
				RoundID:   roundID,
				Type:      "Wager",
				Amount:    floatToDecimal128(wagerAmount),
				Currency:  currency,
				USDAmount: floatToDecimal128(wagerUSD),
			}

			// Payout
			payoutTime := wagerTime.Add(time.Duration(rand.Intn(600)+10) * time.Second)
			payoutAmount := 0.0
			// 45% chance of winning
			if rand.Float64() < 0.45 {
				payoutAmount = wagerAmount * (1.5 + rand.Float64()) // 1.5x to 2.5x win
			}
			payoutUSD := payoutAmount * rates[currency]

			payout := models.Transaction{
				ID:        primitive.NewObjectID(),
				CreatedAt: payoutTime,
				UserID:    userID,
				RoundID:   roundID,
				Type:      "Payout",
				Amount:    floatToDecimal128(payoutAmount),
				Currency:  currency,
				USDAmount: floatToDecimal128(payoutUSD),
			}

			docs = append(docs, wager, payout)
		}

		_, err := coll.InsertMany(context.Background(), docs, options.InsertMany().SetOrdered(false))
		if err != nil {
			log.Warnf("Worker %d: Insert error: %v", workerID, err)
		}

		roundsLeft -= currentBatchRounds
		if roundsLeft%100000 == 0 && roundsLeft != 0 {
			log.Infof("Worker %d: %d rounds remaining...", workerID, roundsLeft)
		}
	}
	log.Infof("Worker %d: Finished", workerID)
}

func floatToDecimal128(f float64) primitive.Decimal128 {
	d, _ := primitive.ParseDecimal128(fmt.Sprintf("%.10f", f))
	return d
}
