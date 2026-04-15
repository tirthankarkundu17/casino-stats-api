package repository

import (
	"context"
	"time"

	"tirthankarkundu17/casino-analytics/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type TransactionRepository interface {
	GetGGR(ctx context.Context, from, to time.Time) ([]models.GGRResult, error)
	GetDailyWagerVolume(ctx context.Context, from, to time.Time) ([]models.DailyVolumeResult, error)
	GetAllUsersWagerRank(ctx context.Context, from, to time.Time) ([]models.UserWagerRank, error)
	GetUserWagerStats(ctx context.Context, userID primitive.ObjectID, from, to time.Time) (*models.UserWagerStats, error)
}

type mongoTransactionRepository struct {
	collection *mongo.Collection
	log        *zap.SugaredLogger
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *mongo.Database, collectionName string, log *zap.SugaredLogger) TransactionRepository {
	return &mongoTransactionRepository{
		collection: db.Collection(collectionName),
		log:        log,
	}
}

// GetGGR calculates the gross gaming revenue
func (r *mongoTransactionRepository) GetGGR(ctx context.Context, from, to time.Time) ([]models.GGRResult, error) {
	pipeline := mongo.Pipeline{
		{{
			Key: "$match", Value: bson.D{
				{Key: "createdAt", Value: bson.D{
					{Key: "$gte", Value: from},
					{Key: "$lte", Value: to},
				}},
				{Key: "type", Value: bson.D{
					{Key: "$in", Value: bson.A{"Wager", "Payout"}},
				}},
			},
		}},
		{{
			Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$currency"},
				{Key: "totalGGR", Value: bson.D{
					{Key: "$sum", Value: bson.D{
						{Key: "$cond", Value: bson.A{
							bson.D{{Key: "$eq", Value: bson.A{"$type", "Wager"}}},
							"$amount", // If Wager, add amount
							bson.D{{Key: "$multiply", Value: bson.A{"$amount", -1}}}, // If Payout, subtract amount
						}},
					}},
				}},
				{Key: "usdAmount", Value: bson.D{
					{Key: "$sum", Value: bson.D{
						{Key: "$cond", Value: bson.A{
							bson.D{{Key: "$eq", Value: bson.A{"$type", "Wager"}}},
							"$usdAmount", // If Wager, add USD amount
							bson.D{{Key: "$multiply", Value: bson.A{"$usdAmount", -1}}}, // If Payout, subtract USD amount
						}},
					}},
				}},
			},
		}},
	}

	queryStart := time.Now()
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	dbExecutionTime := time.Since(queryStart)
	r.log.Infof("[Performance] GetGGR DB Aggregation Query took: %v", dbExecutionTime)

	decodeStart := time.Now()
	var results []models.GGRResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	r.log.Infof("[Performance] GetGGR Cursor Decoding took: %v", time.Since(decodeStart))
	r.log.Infof("[Performance] GetGGR Total Time (DB + Decode): %v", time.Since(queryStart))

	return results, nil
}

// GetDailyWagerVolume calculates the daily wager volume
func (r *mongoTransactionRepository) GetDailyWagerVolume(ctx context.Context, from, to time.Time) ([]models.DailyVolumeResult, error) {
	pipeline := mongo.Pipeline{
		// 1. Filter by timeframe and type
		{{Key: "$match", Value: bson.D{
			{Key: "type", Value: "Wager"},
			{Key: "createdAt", Value: bson.D{
				{Key: "$gte", Value: from},
				{Key: "$lte", Value: to},
			}},
		}}},

		// 2. Group by date and currency
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "date", Value: bson.D{{Key: "$dateTrunc", Value: bson.D{
					{Key: "date", Value: "$createdAt"},
					{Key: "unit", Value: "day"},
				}}}},
				{Key: "currency", Value: "$currency"},
			}},
			{Key: "totalVolume", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
			{Key: "usdVolume", Value: bson.D{{Key: "$sum", Value: "$usdAmount"}}},
		}}},

		// 3. Sort by date and currency
		{{Key: "$sort", Value: bson.D{{Key: "_id.date", Value: 1}, {Key: "_id.currency", Value: 1}}}},
	}

	opts := options.Aggregate().SetAllowDiskUse(true)

	queryStart := time.Now()
	cursor, err := r.collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	dbExecutionTime := time.Since(queryStart)
	r.log.Infof("[Performance] GetDailyWagerVolume DB Aggregation Query took: %v", dbExecutionTime)

	decodeStart := time.Now()
	var results []models.DailyVolumeResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	r.log.Infof("[Performance] GetDailyWagerVolume Cursor Decoding took: %v", time.Since(decodeStart))
	r.log.Infof("[Performance] GetDailyWagerVolume Total Time (DB + Decode): %v", time.Since(queryStart))

	return results, nil
}

// GetAllUsersWagerRank calculates the wager rank of all users
func (r *mongoTransactionRepository) GetAllUsersWagerRank(ctx context.Context, from, to time.Time) ([]models.UserWagerRank, error) {
	pipeline := mongo.Pipeline{
		// 1. Filter by timeframe and type
		{{Key: "$match", Value: bson.D{
			{Key: "type", Value: "Wager"},
			{Key: "createdAt", Value: bson.D{
				{Key: "$gte", Value: from},
				{Key: "$lte", Value: to},
			}},
		}}},

		// 2. Group by user and sum USD
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$userId"},
			{Key: "totalUSD", Value: bson.D{{Key: "$sum", Value: "$usdAmount"}}},
		}}},

		// 3. Sort by total USD
		{{Key: "$sort", Value: bson.D{{Key: "totalUSD", Value: -1}}}},
	}

	opts := options.Aggregate().SetAllowDiskUse(true)

	queryStart := time.Now()
	cursor, err := r.collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	dbExecutionTime := time.Since(queryStart)
	r.log.Infof("[Performance] GetAllUsersWagerRank DB Aggregation Query took: %v", dbExecutionTime)

	decodeStart := time.Now()
	var results []models.UserWagerRank
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	r.log.Infof("[Performance] GetAllUsersWagerRank Cursor Decoding took: %v", time.Since(decodeStart))
	r.log.Infof("[Performance] GetAllUsersWagerRank Total Time (DB + Decode): %v", time.Since(queryStart))

	return results, nil
}

// GetUserWagerStats calculates the wager stats of a user
func (r *mongoTransactionRepository) GetUserWagerStats(ctx context.Context, userID primitive.ObjectID, from, to time.Time) (*models.UserWagerStats, error) {
	pipeline := mongo.Pipeline{
		// 1. Filter by timeframe and type
		{{Key: "$match", Value: bson.D{
			{Key: "type", Value: "Wager"},
			{Key: "createdAt", Value: bson.D{
				{Key: "$gte", Value: from},
				{Key: "$lte", Value: to},
			}},
		}}},

		// 2. Sum USD by user
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$userId"},
			{Key: "totalUSD", Value: bson.D{{Key: "$sum", Value: "$usdAmount"}}},
		}}},

		// 3. Calculate rank and total user count across the entire window
		{{Key: "$setWindowFields", Value: bson.D{
			{Key: "sortBy", Value: bson.D{{Key: "totalUSD", Value: -1}}},
			{Key: "output", Value: bson.D{
				{Key: "rank", Value: bson.D{{Key: "$rank", Value: bson.D{}}}},
				{Key: "totalUsers", Value: bson.D{{Key: "$count", Value: bson.D{}}}},
			}},
		}}},

		// 4. Narrow down to the specific user
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: userID}}}},
	}

	opts := options.Aggregate().SetAllowDiskUse(true)

	queryStart := time.Now()
	cursor, err := r.collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	r.log.Infof("[Performance] GetUserWagerStats DB Query took: %v", time.Since(queryStart))

	var results []models.UserWagerStats
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil // Not found
	}

	return &results[0], nil
}
