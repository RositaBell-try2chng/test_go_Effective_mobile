package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"subscription-aggregator/internal/models"
	"subscription-aggregator/internal/validation"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type SubscriptionHandler struct {
	db     *sqlx.DB
	logger *logrus.Logger
}

func NewSubscriptionHandler(db *sqlx.DB, logger *logrus.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		db:     db,
		logger: logger,
	}
}

// POST /subscriptions
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSubscriptionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validation.ValidateCreateSubscription(req); err != nil {
		h.logger.WithError(err).Error("Validation failed")
		if validationErrors, ok := err.(validation.ValidationErrors); ok {
			var errorMessages []string
			for _, validationErr := range validationErrors {
				errorMessages = append(errorMessages, validationErr.Message)
			}
			http.Error(w, strings.Join(errorMessages, "; "), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Invalid user ID format")
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Parse start date
	startDate, err := validation.ParseMonthYear(req.StartDate)
	if err != nil {
		h.logger.WithError(err).Error("Invalid start date format")
		http.Error(w, "Invalid start date format (use MM-YYYY)", http.StatusBadRequest)
		return
	}

	// Parse end date
	var endDate *time.Time
	if req.EndDate != "" {
		parsedEndDate, err := validation.ParseMonthYear(req.EndDate)
		if err != nil {
			h.logger.WithError(err).Error("Invalid end date format")
			http.Error(w, "Invalid end date format (use MM-YYYY)", http.StatusBadRequest)
			return
		}
		endDate = &parsedEndDate
	}

	// Insert subscription
	var subscription models.Subscription
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at`

	err = h.db.QueryRowx(query, req.ServiceName, req.Price, userID, startDate, endDate).StructScan(&subscription)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create subscription")
		http.Error(w, "Failed to create subscription", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subscription)

	h.logger.WithFields(logrus.Fields{
		"subscription_id": subscription.ID,
		"user_id":         subscription.UserID,
		"service_name":    subscription.ServiceName,
	}).Info("Subscription created successfully")
}

// GET /subscriptions/{id}
func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/subscriptions/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid subscription ID format")
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	var subscription models.Subscription
	query := `SELECT * FROM subscriptions WHERE id = $1`
	err = h.db.Get(&subscription, query, id)
	if err != nil {
		h.logger.WithError(err).Error("Subscription not found")
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscription)
}

// PUT /subscriptions/{id}
func (h *SubscriptionHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/subscriptions/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid subscription ID format")
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate
	if err := validation.ValidateUpdateSubscription(req); err != nil {
		h.logger.WithError(err).Error("Validation failed")
		if validationErrors, ok := err.(validation.ValidationErrors); ok {
			var errorMessages []string
			for _, validationErr := range validationErrors {
				errorMessages = append(errorMessages, validationErr.Message)
			}
			http.Error(w, strings.Join(errorMessages, "; "), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	// Build update query
	setParts := []string{}
	args := []interface{}{}
	argCount := 1

	if req.ServiceName != "" {
		setParts = append(setParts, fmt.Sprintf("service_name = $%d", argCount))
		args = append(args, req.ServiceName)
		argCount++
	}

	if req.Price > 0 {
		setParts = append(setParts, fmt.Sprintf("price = $%d", argCount))
		args = append(args, req.Price)
		argCount++
	}

	if req.StartDate != "" {
		startDate, err := validation.ParseMonthYear(req.StartDate)
		if err != nil {
			h.logger.WithError(err).Error("Invalid start date format")
			http.Error(w, "Invalid start date format (use MM-YYYY)", http.StatusBadRequest)
			return
		}
		setParts = append(setParts, fmt.Sprintf("start_date = $%d", argCount))
		args = append(args, startDate)
		argCount++
	}

	if req.EndDate != "" {
		if req.EndDate == "null" {
			setParts = append(setParts, fmt.Sprintf("end_date = $%d", argCount))
			args = append(args, nil)
		} else {
			endDate, err := validation.ParseMonthYear(req.EndDate)
			if err != nil {
				h.logger.WithError(err).Error("Invalid end date format")
				http.Error(w, "Invalid end date format (use MM-YYYY)", http.StatusBadRequest)
				return
			}
			setParts = append(setParts, fmt.Sprintf("end_date = $%d", argCount))
			args = append(args, endDate)
		}
		argCount++
	}

	if len(setParts) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())
	argCount++

	args = append(args, id)
	query := fmt.Sprintf("UPDATE subscriptions SET %s WHERE id = $%d", strings.Join(setParts, ", "), argCount)

	result, err := h.db.Exec(query, args...)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update subscription")
		http.Error(w, "Failed to update subscription", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get rows affected")
		http.Error(w, "Failed to update subscription", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	h.logger.WithField("subscription_id", id).Info("Subscription updated successfully")
}

// DELETE /subscriptions/{id}
func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/subscriptions/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid subscription ID format")
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec("DELETE FROM subscriptions WHERE id = $1", id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete subscription")
		http.Error(w, "Failed to delete subscription", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get rows affected")
		http.Error(w, "Failed to delete subscription", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	h.logger.WithField("subscription_id", id).Info("Subscription deleted successfully")
}

// GET /subscriptions
func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	query := "SELECT * FROM subscriptions"
	conditions := []string{}
	args := []interface{}{}
	argCount := 0

	// Parse query parameters
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.logger.WithError(err).Error("Invalid user ID format")
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}
		argCount++
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argCount))
		args = append(args, userID)
	}

	if serviceName := r.URL.Query().Get("service_name"); serviceName != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("service_name ILIKE $%d", argCount))
		args = append(args, "%"+serviceName+"%")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add pagination
	limit := 100 // default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}
	query += fmt.Sprintf(" LIMIT %d", limit)

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
			query += fmt.Sprintf(" OFFSET %d", offset)
		}
	}

	var subscriptions []models.Subscription
	err := h.db.Select(&subscriptions, query, args...)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list subscriptions")
		http.Error(w, "Failed to list subscriptions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscriptions)
}

// POST /subscriptions/aggregate
func (h *SubscriptionHandler) AggregateSubscriptions(w http.ResponseWriter, r *http.Request) {
	var req models.AggregationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validation.ValidateAggregationRequest(req); err != nil {
		h.logger.WithError(err).Error("Validation failed")
		if validationErrors, ok := err.(validation.ValidationErrors); ok {
			var errorMessages []string
			for _, validationErr := range validationErrors {
				errorMessages = append(errorMessages, validationErr.Message)
			}
			http.Error(w, strings.Join(errorMessages, "; "), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	// Parse dates
	startDate, err := validation.ParseMonthYear(req.StartDate)
	if err != nil {
		h.logger.WithError(err).Error("Invalid start date format")
		http.Error(w, "Invalid start date format (use MM-YYYY)", http.StatusBadRequest)
		return
	}

	endDate, err := validation.ParseMonthYear(req.EndDate)
	if err != nil {
		h.logger.WithError(err).Error("Invalid end date format")
		http.Error(w, "Invalid end date format (use MM-YYYY)", http.StatusBadRequest)
		return
	}

	// Build query
	query := `
		SELECT COALESCE(SUM(price), 0) as total_cost
		FROM subscriptions
		WHERE start_date >= $1 AND (end_date IS NULL OR end_date <= $2)`

	args := []interface{}{startDate, endDate}

	argCount := 2
	if req.UserID != nil {
		argCount++
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, req.UserID)
	}

	if req.ServiceName != nil {
		argCount++
		query += fmt.Sprintf(" AND service_name ILIKE $%d", argCount)
		args = append(args, "%"+*req.ServiceName+"%")
	}

	var result struct {
		TotalCost int64 `db:"total_cost"`
	}

	err = h.db.Get(&result, query, args...)
	if err != nil {
		h.logger.WithError(err).Error("Failed to aggregate subscriptions")
		http.Error(w, "Failed to aggregate subscriptions", http.StatusInternalServerError)
		return
	}

	response := models.AggregationResponse{
		TotalCost: result.TotalCost,
		Period:    fmt.Sprintf("%s to %s", req.StartDate, req.EndDate),
		UserID:    req.UserID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.WithFields(logrus.Fields{
		"total_cost": result.TotalCost,
		"period":     response.Period,
	}).Info("Subscription aggregation completed")
}


