package routes

import (
	"net/http"

	"tirthankarkundu17/casino-stats-api/internal/handlers"
	"tirthankarkundu17/casino-stats-api/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, analyticsHandler *handlers.AnalyticsHandler, authToken string) {
	// Public health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Private routes with Auth Middleware
	api := r.Group("/")
	api.Use(middlewares.AuthMiddleware(authToken))
	{
		api.GET("/gross_gaming_rev", analyticsHandler.GetGGR)
		api.GET("/daily_wager_volume", analyticsHandler.GetDailyWagerVolume)

		// User group
		user := api.Group("/user")
		{
			user.GET("/:user_id/wager_percentile", analyticsHandler.GetUserWagerPercentile)
		}
	}
}
