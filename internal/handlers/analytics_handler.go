package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"tirthankarkundu17/casino-analytics/internal/cache"
	"tirthankarkundu17/casino-analytics/internal/models"
	"tirthankarkundu17/casino-analytics/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

const DefaultCacheTTL = 5 * time.Minute

type AnalyticsHandler struct {
	service *services.AnalyticsService
	cache   *cache.Cache
	log     *zap.SugaredLogger
}

func NewAnalyticsHandler(s *services.AnalyticsService, c *cache.Cache, log *zap.SugaredLogger) *AnalyticsHandler {
	return &AnalyticsHandler{service: s, cache: c, log: log}
}

// handleServiceError maps typed service errors to appropriate HTTP status codes.
func (h *AnalyticsHandler) handleServiceError(c *gin.Context, err error) {
	var notFoundErr *models.ErrNotFound
	var noDataErr *models.ErrNoData

	switch {
	case errors.As(err, &notFoundErr):
		h.sendError(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", notFoundErr.Error(), nil)
	case errors.As(err, &noDataErr):
		h.sendError(c, http.StatusNotFound, "NO_DATA_FOUND", noDataErr.Error(), nil)
	default:
		h.log.Errorf("Internal error: %v", err)
		h.sendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred", nil)
	}
}

// sendError is a helper to transmit structured errors
func (h *AnalyticsHandler) sendError(c *gin.Context, status int, code, msg string, details []models.ErrorDetail) {
	c.JSON(status, gin.H{
		"success": false,
		"error": models.APIError{
			Code:    code,
			Message: msg,
			Details: details,
		},
	})
}

// bindAndValidateTimeframe binds query params and runs business validation.
func (h *AnalyticsHandler) bindAndValidateTimeframe(c *gin.Context) (models.TimeframeParams, bool) {
	var params models.TimeframeParams
	if err := c.ShouldBindQuery(&params); err != nil {
		var details []models.ErrorDetail
		var vErrs validator.ValidationErrors
		if errors.As(err, &vErrs) {
			for _, f := range vErrs {
				details = append(details, models.ErrorDetail{
					Field: f.Field(),
					Issue: fmt.Sprintf("failed on %s validation", f.Tag()),
				})
			}
		}
		h.sendError(c, http.StatusBadRequest, "VALIDATION_FAILED", "Invalid request parameters", details)
		return params, false
	}

	if err := params.Validate(); err != nil {
		details := []models.ErrorDetail{{Field: "timeframe", Issue: err.Error()}}
		h.sendError(c, http.StatusBadRequest, "BUSINESS_RULE_VIOLATION", "Logical validation failed", details)
		return params, false
	}

	return params, true
}

// sendSuccess is a helper to transmit successful data
func (h *AnalyticsHandler) sendSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetGGR calculates the gross gaming revenue
func (h *AnalyticsHandler) GetGGR(c *gin.Context) {
	params, ok := h.bindAndValidateTimeframe(c)
	if !ok {
		return
	}

	cacheKey := fmt.Sprintf("ggr:%s:%s", params.From.Format("2006-01-02"), params.To.Format("2006-01-02"))
	var result any
	if err := h.cache.Get(c.Request.Context(), cacheKey, &result); err == nil {
		h.sendSuccess(c, result)
		return
	}

	res, err := h.service.GetGGR(c.Request.Context(), params.From, params.To)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.cache.Set(c.Request.Context(), cacheKey, res, DefaultCacheTTL)
	h.sendSuccess(c, res)
}

// GetDailyWagerVolume calculates the daily wager volume
func (h *AnalyticsHandler) GetDailyWagerVolume(c *gin.Context) {
	params, ok := h.bindAndValidateTimeframe(c)
	if !ok {
		return
	}

	cacheKey := fmt.Sprintf("daily_volume:%s:%s", params.From.Format("2006-01-02"), params.To.Format("2006-01-02"))
	var result any
	if err := h.cache.Get(c.Request.Context(), cacheKey, &result); err == nil {
		h.sendSuccess(c, result)
		return
	}

	res, err := h.service.GetDailyWagerVolume(c.Request.Context(), params.From, params.To)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.cache.Set(c.Request.Context(), cacheKey, res, DefaultCacheTTL)
	h.sendSuccess(c, res)
}

// GetUserWagerPercentile calculates the wager percentile of a user
func (h *AnalyticsHandler) GetUserWagerPercentile(c *gin.Context) {
	params, ok := h.bindAndValidateTimeframe(c)
	if !ok {
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format", []models.ErrorDetail{{Field: "user_id", Issue: "Must be a valid hex string"}})
		return
	}

	cacheKey := fmt.Sprintf("percentile:%s:%s:%s", userIDStr, params.From.Format("2006-01-02"), params.To.Format("2006-01-02"))
	var result any
	if err := h.cache.Get(c.Request.Context(), cacheKey, &result); err == nil {
		h.sendSuccess(c, result)
		return
	}

	res, err := h.service.GetUserWagerPercentile(c.Request.Context(), userID, params.From, params.To)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.cache.Set(c.Request.Context(), cacheKey, res, DefaultCacheTTL)
	h.sendSuccess(c, res)
}
